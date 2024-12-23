package model

type User struct {
    Id int `json:"id"`
    Name string `json:"name"`
    UserName string `json:"user_name"`
    ChatId string `json:"chat_id"`
    Role string `json:"role"`
    Portfolio string  `json:"portfolio"`
    Specialization string `json:"specialization"`
}
