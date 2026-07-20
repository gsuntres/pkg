package persist

const USER_COLLECTION = "user"

// User can use [User.EncryptPassword] to encrypt the password
type User struct {
	Id          string `bson:"_id" json:"id"`
	Fullname 		string `bson:"fullname" json:"fullname"`
	Firstname   string `bson:"firstname" json:"firstname"`
	Username		string `bson:"username" json:"username"`
	Password		string `bson:"password" json:"password"`
	Email    		string `bson:"email" json:"email"`
	Roles       []string `bson:"roles" json:"roles"`
	Credentials map[string]any `bson:"credentials" json:"credentials"`
}

const ACCOUNT_COLLECTION = "account"
// Deprecated: use core.BusinessAccount
type Account struct {
	Id          string `bson:"_id" json:"id"`
	Code        string `bson:"code" json:"code"`
	Owner       string `bson:"owner" json:"owner"`
}
