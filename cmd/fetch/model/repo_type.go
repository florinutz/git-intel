package model

// RepositoryType represents the type of repository.
type RepositoryType int

const (
	RepoTypeAll RepositoryType = iota
	RepoTypePublic
	RepoTypePrivate
	RepoTypeForks
	RepoTypeSources
	RepoTypeMember
)

// String returns the string representation of a RepositoryType.
// The returned string is based on the value of the RepositoryType
// and is determined by the map that associates RepositoryType values
// with their corresponding string representations.
func (r RepositoryType) String() string {
	return map[RepositoryType]string{
		RepoTypeAll:     "all",
		RepoTypePublic:  "public",
		RepoTypePrivate: "private",
		RepoTypeForks:   "forks",
		RepoTypeSources: "sources",
		RepoTypeMember:  "member",
	}[r]
}
