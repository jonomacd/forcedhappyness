package certificate

import (
	"os"
	"os/exec"
)

func Run(email string) error {

	// err := execute("certbot", "certonly", "--standalone", "--agree-tos", "--email", email, "-n", "-d", "iwillbenice.com", "-d", "www.iwillbenice.com")
	// if err != nil {
	// 	return err
	// }

	return nil

}

func execute(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
