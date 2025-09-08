package httpapi

type SignUpIn struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"  binding:"required"`
	Email     string `json:"email"      binding:"required,email"`
	Password1 string `json:"password1"  binding:"required,min=8"`
	Password2 string `json:"password2"  binding:"required,min=8"`
	City      string `json:"city"       binding:"required"`
	Country   string `json:"country"    binding:"required"`
}

type LoginIn struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
