package utils

var (
	DefaultCategory = "Not Processed"
	JWTSecret       = []byte("JWT_Secret")
)

var (
	// Headers are the first row in the exported xlsx/CSV file
	Headers = [...]string{
		"Identification Number",
		"Mode of attendance code",
		"Type of attendance code",
		"NESB indicator",
		"Citizenship indicator",
		"Study area",
		"RawComment",
		"TrainCategory",
		"ClassifiedCategory",
		"ClassifiedConfidence",
		"HumanCategory",
		"FinalClassification",
	}
)

// Unified definition of roles in this system
const (
	RoleModerator = "Moderator"
	RoleAdmin     = "Admin"
)
