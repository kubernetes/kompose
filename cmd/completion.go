package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var completion = &cobra.Command{
	Use:   "completion SHELL",
	Short: "Output shell completion code",
	Long: `Generates shell completion code.

Auto completion supports both bash and zsh. Output is to STDOUT.

source <(kompose completion bash)
source <(kompose completion zsh)

Will load the shell completion code.
	`,

	RunE: func(cmd *cobra.Command, args []string) error {

		err := Generate(cmd, args)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}

		return nil
	},
}

// Generate the appropriate autocompletion file
func Generate(cmd *cobra.Command, args []string) error {

	// Check the passed in arguments
	if len(args) == 0 {
		return fmt.Errorf("Shell not specified. ex. kompose completion [bash|zsh]")
	}
	if len(args) > 1 {
		return fmt.Errorf("Too many arguments. Expected only the shell type. ex. kompose completion [bash|zsh]")
	}
	shell := args[0]

	// Generate bash through cobra if selected
	if shell == "bash" {
		return cmd.Root().GenBashCompletion(os.Stdout)

		// Generate zsh with the appropriate conversion as well as bash inclusion
	} else if shell == "zsh" {
		return runCompletionZsh(os.Stdout, cmd.Root())

		// Else, return an error.
	} else {
		return fmt.Errorf("not a compatible shell, bash and zsh are only supported")
	}
}

func init() {
	RootCmd.AddCommand(completion)
}

/*
	This is copied from
	https://github.com/kubernetes/kubernetes/blob/ea18d5c32ee7c320fe96dda6b0c757476908e696/pkg/kubectl/cmd/completion.go
	in order to generate ZSH completion support.
*/
func runCompletionZsh(out io.Writer, kompose *cobra.Command) error {

	zshInitialization := `
__kompose_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand
	source "$@"
}
__kompose_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift
		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__kompose_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}
__kompose_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?
	# filter by given word as prefix
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			echo "${w}"
		fi
	done
}
__kompose_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}
__kompose_declare() {
	if [ "$1" == "-F" ]; then
		whence -w "$@"
	else
		builtin declare "$@"
	fi
}
__kompose_ltrim_colon_completions()
{
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}
__kompose_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}
__kompose_filedir() {
	local RET OLD_IFS w qw
	__debug "_filedir $@ cur=$cur"
	if [[ "$1" = \~* ]]; then
		# somehow does not work. Maybe, zsh does not call this at all
		eval echo "$1"
		return 0
	fi
	OLD_IFS="$IFS"
	IFS=$'\n'
	if [ "$1" = "-d" ]; then
		shift
		RET=( $(compgen -d) )
	else
		RET=( $(compgen -f) )
	fi
	IFS="$OLD_IFS"
	IFS="," __debug "RET=${RET[@]} len=${#RET[@]}"
	for w in ${RET[@]}; do
		if [[ ! "${w}" = "${cur}"* ]]; then
			continue
		fi
		if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
			qw="$(__kompose_quote "${w}")"
			if [ -d "${w}" ]; then
				COMPREPLY+=("${qw}/")
			else
				COMPREPLY+=("${qw}")
			fi
		fi
	done
}
__kompose_quote() {
    if [[ $1 == \'* || $1 == \"* ]]; then
        # Leave out first character
        printf %q "${1:1}"
    else
    	printf %q "$1"
    fi
}
autoload -U +X bashcompinit && bashcompinit
# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q GNU; then
	LWORD='\<'
	RWORD='\>'
fi
__kompose_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__kompose_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__kompose_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__kompose_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__kompose_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__kompose_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/__kompose_declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__kompose_type/g" \
	<<'BASH_COMPLETION_EOF'
`
	out.Write([]byte(zshInitialization))

	buf := new(bytes.Buffer)
	kompose.GenBashCompletion(buf)
	out.Write(buf.Bytes())

	zshTail := `
BASH_COMPLETION_EOF
}
__kompose_bash_source <(__kompose_convert_bash_to_zsh)
`
	out.Write([]byte(zshTail))
	return nil
}
