package auth

type User struct {
    ID       uint
    Email    string
    Password Password
}

func (u *User) SetID(id uint){
    u.ID = id
}