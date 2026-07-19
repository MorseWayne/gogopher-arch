package releaseplan

import "errors"

type Config struct {
	RuntimeUser   string
	Checks        []string
	MigrationMode string
}

func Validate(config Config) error {
	return errors.New("TODO: validate release gates")
}
