package shared

// User is the sample entity used by the kvx repository examples.
type User struct {
	ID    string `kvx:"id"`
	Name  string `kvx:"name"`
	Email string `kvx:"email,index=email"`
}
