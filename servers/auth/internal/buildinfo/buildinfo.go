package buildinfo

// Version é definida no build com:
//
//	go build -ldflags "-X auth/internal/buildinfo.Version=1.0.0-rc.1"
//
// Sem a flag, fica "dev".
var Version = "dev"
