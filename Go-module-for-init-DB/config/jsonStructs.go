package jsonModal

type Config struct{
    Env         string  `json:"environment"`
    Port        string  `json:"port"`
}
type Machine struct {
    MariaDB     MariaDbConfig `json:"mariadb"`
}

type MariaDbConfig struct {
    InUse       bool   `json:"in_use"`
    User        string `json:"user"`
    Password    string `json:"password"`
    Host        string `json:"host"`
    Port        string `json:"port"`
    Name        string `json:"name"`
}
