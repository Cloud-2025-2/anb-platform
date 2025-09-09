#!/bin/bash

# Kafka Monitoring Script for ANB Platform
# This script provides real-time monitoring of Kafka topics and consumer groups

echo "📊 Kafka Monitoring Dashboard"
echo "============================="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Function to check if Kafka is running
check_kafka() {
    if docker exec kafka kafka-broker-api-versions --bootstrap-server localhost:9092 >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Kafka is running${NC}"
        return 0
    else
        echo -e "${RED}❌ Kafka is not accessible${NC}"
        return 1
    fi
}

# Function to list topics
list_topics() {
    echo -e "\n${BLUE}📝 Kafka Topics${NC}"
    echo "==============="
    
    topics=$(docker exec kafka kafka-topics --bootstrap-server localhost:9092 --list 2>/dev/null)
    
    if [ -z "$topics" ]; then
        echo -e "${YELLOW}No topics found${NC}"
    else
        echo "$topics" | while read topic; do
            if [[ $topic == video-processing* ]]; then
                echo -e "${GREEN}📹 $topic${NC}"
            else
                echo -e "📄 $topic"
            fi
        done
    fi
}

# Function to show topic details
show_topic_details() {
    local topic=$1
    echo -e "\n${BLUE}📊 Topic Details: $topic${NC}"
    echo "================================"
    
    # Get topic description
    docker exec kafka kafka-topics --bootstrap-server localhost:9092 --describe --topic "$topic" 2>/dev/null
    
    # Get message count (approximate)
    echo -e "\n${YELLOW}Message Count (approximate):${NC}"
    docker exec kafka kafka-run-class kafka.tools.GetOffsetShell \
        --broker-list localhost:9092 --topic "$topic" --time -1 2>/dev/null | \
        awk -F: '{sum += $3} END {print sum}'
}

# Function to monitor consumer groups
monitor_consumer_groups() {
    echo -e "\n${BLUE}👥 Consumer Groups${NC}"
    echo "=================="
    
    groups=$(docker exec kafka kafka-consumer-groups --bootstrap-server localhost:9092 --list 2>/dev/null)
    
    if [ -z "$groups" ]; then
        echo -e "${YELLOW}No consumer groups found${NC}"
        return
    fi
    
    echo "$groups" | while read group; do
        echo -e "\n${GREEN}Group: $group${NC}"
        docker exec kafka kafka-consumer-groups --bootstrap-server localhost:9092 \
            --describe --group "$group" 2>/dev/null | head -10
    done
}

# Function to show real-time message flow
monitor_messages() {
    local topic=$1
    echo -e "\n${BLUE}🔄 Real-time Messages: $topic${NC}"
    echo "=================================="
    echo "Press Ctrl+C to stop monitoring"
    echo ""
    
    docker exec -it kafka kafka-console-consumer \
        --bootstrap-server localhost:9092 \
        --topic "$topic" \
        --from-beginning \
        --max-messages 10 2>/dev/null || true
}

# Function to publish test message
publish_test_message() {
    local topic=$1
    local message=$2
    
    echo -e "\n${BLUE}📤 Publishing test message to $topic${NC}"
    echo "$message" | docker exec -i kafka kafka-console-producer \
        --bootstrap-server localhost:9092 \
        --topic "$topic" 2>/dev/null
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Message published successfully${NC}"
    else
        echo -e "${RED}❌ Failed to publish message${NC}"
    fi
}

# Main menu
show_menu() {
    echo -e "\n${BLUE}🎛️  Kafka Monitoring Options${NC}"
    echo "============================"
    echo "1. Check Kafka status"
    echo "2. List all topics"
    echo "3. Show video-processing topic details"
    echo "4. Show video-processing-retry topic details"
    echo "5. Show video-processing-dlq topic details"
    echo "6. Monitor consumer groups"
    echo "7. Monitor video-processing messages (real-time)"
    echo "8. Publish test message to video-processing"
    echo "9. Show all topic statistics"
    echo "0. Exit"
    echo ""
    read -p "Select option (0-9): " choice
}

# Statistics function
show_statistics() {
    echo -e "\n${BLUE}📈 Kafka Statistics${NC}"
    echo "==================="
    
    topics=("video-processing" "video-processing-retry" "video-processing-dlq")
    
    for topic in "${topics[@]}"; do
        echo -e "\n${YELLOW}Topic: $topic${NC}"
        
        # Check if topic exists
        if docker exec kafka kafka-topics --bootstrap-server localhost:9092 --list 2>/dev/null | grep -q "^$topic$"; then
            # Get partition count
            partitions=$(docker exec kafka kafka-topics --bootstrap-server localhost:9092 --describe --topic "$topic" 2>/dev/null | grep "PartitionCount" | awk '{print $2}')
            echo "  Partitions: $partitions"
            
            # Get message count
            msg_count=$(docker exec kafka kafka-run-class kafka.tools.GetOffsetShell \
                --broker-list localhost:9092 --topic "$topic" --time -1 2>/dev/null | \
                awk -F: '{sum += $3} END {print sum}')
            echo "  Messages: ${msg_count:-0}"
        else
            echo -e "  ${YELLOW}Topic not found (will be created on first use)${NC}"
        fi
    done
}

# Main loop
while true; do
    show_menu
    
    case $choice in
        1)
            check_kafka
            ;;
        2)
            list_topics
            ;;
        3)
            show_topic_details "video-processing"
            ;;
        4)
            show_topic_details "video-processing-retry"
            ;;
        5)
            show_topic_details "video-processing-dlq"
            ;;
        6)
            monitor_consumer_groups
            ;;
        7)
            monitor_messages "video-processing"
            ;;
        8)
            test_message='{"video_id":"test-123","user_id":"test-user","timestamp":"'$(date -Iseconds)'"}'
            publish_test_message "video-processing" "$test_message"
            ;;
        9)
            show_statistics
            ;;
        0)
            echo -e "${GREEN}👋 Goodbye!${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}Invalid option. Please try again.${NC}"
            ;;
    esac
    
    echo ""
    read -p "Press Enter to continue..."
done
