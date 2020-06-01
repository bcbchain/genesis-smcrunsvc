package version

var (
	// GitCommit is the current HEAD set using ldflags.
	GitCommit string

	// Version is the built softwares version.
	Version = ContractSemVer
)

func init() {
	if GitCommit != "" {
		Version += "-" + GitCommit
	}
}

const (
	// ContractSemVer is the current version of contract.
	// It's the Semantic Version of the software.
	// Must be a string because scripts like dist.sh read this file.
	// XXX: Don't change the name of this variable or you will break
	// automation :)
	ContractSemVer = "0.0.1"
)
