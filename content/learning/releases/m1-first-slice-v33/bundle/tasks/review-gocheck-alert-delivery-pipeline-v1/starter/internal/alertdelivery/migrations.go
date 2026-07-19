package alertdelivery

import "errors"

type Migration struct{ Name, SHA256 string }

func ValidateManifest(entries []Migration) error {
	return errors.New("TODO: validate migration manifest")
}
