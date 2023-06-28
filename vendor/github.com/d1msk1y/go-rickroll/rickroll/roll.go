package rickroll

import (
	"fmt"
	"os/exec"
	"runtime"
)

func RickRoll() {
	url := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	fmt.Println("Never gonna give u up!")

	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", url).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	default:
		fmt.Errorf("unsupported platform")
	}

}
