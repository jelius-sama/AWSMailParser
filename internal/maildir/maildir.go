package maildir

import (
	"AWSMailParser/vars"
	"fmt"
	"github.com/jelius-sama/logger"
	"os"
	"os/exec"
)

func EnsureMaildir(domain, local string) error {
	path := fmt.Sprintf("%s/%s/%s/Maildir", vars.MaildirBase, domain, local)

	// Check if path already exists
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	// Create the maildir structure
	for _, subdir := range []string{"new", "cur", "tmp"} {
		subdirPath := fmt.Sprintf("%s/%s", path, subdir)
		if err := os.MkdirAll(subdirPath, 0700); err != nil {
			logger.TimedError("Failed to create Maildir", path, ":", err)
			return err
		}
	}

	// Set ownership and permissions
	userPath := fmt.Sprintf("%s/%s/%s", vars.MaildirBase, domain, local)

	if err := exec.Command("chown", "-R", "vmail:mail", userPath).Run(); err != nil {
		logger.TimedError("Failed to set ownership for", userPath, ":", err)
	}

	if err := exec.Command("chmod", "-R", "700", userPath).Run(); err != nil {
		logger.TimedError("Failed to set permissions for", userPath, ":", err)
	}

	logger.TimedOkay("Created Maildir:", path)
	return nil
}
