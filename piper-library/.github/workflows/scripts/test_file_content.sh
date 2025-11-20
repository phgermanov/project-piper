# shellcheck disable=SC2086

args=("$@")
# first argument. This must be path to file to test
file_name=${args[0]}
# all other arguments must be strings to search in file
strings_to_check=("${args[@]:1}")

# grep will have exit_code=0 when pattern/string found in file, otherwise exit_code=1
grep_exit_code_sum=0
for str in "${strings_to_check[@]}";
do
  ((grep_exit_code_sum+=$(grep -Eq $str $file_name; echo $?)))
done

if [[ $grep_exit_code_sum != 0 ]]; then
  echo "$file_name file is malformed"
  exit 1
fi

echo "$file_name file test is successful"

#test_file_content.sh piper-defaults-github.yml general: stages: steps:
#test_file_content.sh piper-stage-config.yml apiVersion: metadata: spec:
