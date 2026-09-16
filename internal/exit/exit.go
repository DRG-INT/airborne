package exit

type Code int

const (
	Success              Code = 0
	ExecutionFailure     Code = 1
	InvalidInvocation    Code = 2
	VerificationFailure  Code = 3
	ProviderFailure      Code = 4
	PolicyContextFailure Code = 5
)

var messages = map[Code]string{
	Success:              "success",
	ExecutionFailure:     "execution failure",
	InvalidInvocation:    "invalid invocation",
	VerificationFailure:  "verification failure",
	ProviderFailure:      "provider failure",
	PolicyContextFailure: "policy/context failure",
}

func (c Code) String() string {
	return messages[c]
}
