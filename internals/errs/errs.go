package errs

import "errors"

type Error struct {
	Message string
	Err     error
}

var (
	// File System Errors
	ErrFileNotFound = errors.New("file not found")
	ErrDirNotFound  = errors.New("directory not found")
	ErrInvalidPath  = errors.New("invalid path")

	// Toml Error
	ErrWritingToml  = errors.New("error writing toml file")
	ErrDecodingToml = errors.New("error decoding toml file")
	ErrFixingToml   = errors.New("error fixing toml file")

	// Identity
	ErrCreatingKeys      = errors.New("error creating keys")
	ErrInvalidPublicKey  = errors.New("invalid public key")
	ErrInvalidPrivateKey = errors.New("invalid private key")
	ErrMismatchedKeyPair = errors.New("public and private keys do not match")
	ErrMissingKeys       = errors.New("public and private keys are required")

	// Wire
	ErrDefault           = errors.New("error processing stream")
	ErrUnknowStreamType  = errors.New("unknow stream type")
	ErrStreamToLarge     = errors.New("stream to large")
	ErrStreamInfoToLarge = errors.New("stream info to large")
	ErrStreamSizeZero    = errors.New("stream size zero")
	ErrInvalidStreamInfo = errors.New("invalid stream info")
)

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Is(target error) bool {
	return errors.Is(e.Err, target)
}

func NewError(err error, message string) error {
	return &Error{
		Message: message,
		Err:     err,
	}
}
