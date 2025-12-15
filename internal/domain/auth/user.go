package auth

type User struct {
    ID       uint
    Email    string
    Password Password
}