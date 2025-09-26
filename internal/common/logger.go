package common
import (
	"log"
	"os"
)

var Logger *log.Logger

func SetupLogger() {
	Logger = log.New(os.Stdout, "[APP] ", log.LstdFlags)
}
