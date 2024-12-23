package model

import "time"

type Order struct {
    ID int 
    Title string
    Description string
    Location string
    User User
    Specialization string
    CreatedAt time.Time
}
