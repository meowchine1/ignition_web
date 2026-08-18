package db

import (

	"time"
)

type OSType string

const (
	OSWindows OSType = "windows"
// 	OSLinux   OSType = "linux"
  )

type FirmwareRecord struct {
	ID        int64
	Name      string
	Size      int64 
	SHA256    string
	Path      string
	IsAvailable bool
	CreatedAt time.Time
}

type FlasherRecord struct {
	ID        int64
	Name      string
	OS        string 
	Size      int64
	SHA256    string
	Path      string
	IsCurrent bool
	CreatedAt time.Time
}

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`  
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}