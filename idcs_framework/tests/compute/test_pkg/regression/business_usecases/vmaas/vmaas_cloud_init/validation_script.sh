#!/bin/bash

# Function to check if a command was successfully executed
check_command() {
    local command=$1

    if ! $command; then
        echo "Command '$command' failed to execute"
        missing=true
    else
        echo "Command '$command' executed successfully"
    fi
}

# Function to check if a public key is present
check_key() {
    local expected_key=$1
    local key_path=$2

    if ! sudo grep -qF "$expected_key" "$key_path"; then
        echo "The expected public key is not present in $key_path"
        missing=true
    else
        echo "The expected public key is present in $key_path"
    fi
}

# Function to check if a file exists with specified content and append flag
check_file() {
    local file_path=$1
    local content=$2
    local append_flag=$3

    if [ ! -f "$file_path" ]; then
        echo "$file_path: Not Found"
        missing=true
    elif ! grep -qF "$content" "$file_path"; then
        echo "$file_path: Content not found"
        missing=true
    elif [ "$append_flag" != "true" ]; then
        echo "$file_path: Append flag is not set to true"
        missing=true
    else
        echo "$file_path: Found with correct content and append flag"
    fi
}

# Flag to indicate if anything is failed
missing=false

# List of packages to check
packages="qemu-guest-agent socat ipset conntrack"

# Sleep to ensure packages are installed 
sleep 30

for pkg in $packages; do
    if ! dpkg -s $pkg | grep -q "Status: install ok installed"; then
        echo "$pkg: Not Installed"
        missing=true
    else
        echo "$pkg: Installed"
    fi
done


# Check for the presence and permissions of /etc/helloworld3 file
if [ ! -f "/etc/helloworld3" ]; then
    echo "/etc/helloworld3: Not Found"
    missing=true
elif [ "$(stat -c "%a" /etc/helloworld3)" != "700" ]; then
    echo "/etc/helloworld3: Incorrect Permissions"
    missing=true
else
    echo "/etc/helloworld3: Found with Correct Permissions"
fi


# Check for the presence and difference in permissions of /etc/helloworld file
if [ ! -f "/etc/helloworld" ]; then
    echo "/etc/helloworld: Not Found"
    missing=true
elif [ "$(stat -c "%a" /etc/helloworld)" != "777" ]; then
    echo "/etc/helloworld: Incorrect Permissions"
    missing=true
else
    echo "/etc/helloworld: Found with Correct Permissions"
fi

# Check for the presence of the specific public key in authorized_keys
expected_key1="ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCcO4pgGXXz/IDOOk0QcK4n45bkwfhr8TgCLLN1e2Qm5Zpda6egvpeI+ZrYpYNnEvCeIrZHFjwCL0JvkqY2xvf0EF5BiOa1dWc/eDj9csgW0xihQUETcUbgEQDCy8Ph3t7DGqw+h5yh6CwT2oIe9jcNnQmd+097X8aYvxk3zVx8/E7QqBmUDUH23U1VDdHOiB4ie+QUsUrmsKVxI3zZhpDvToY7maRS2TfJe0wucGrGrpvqOx+YF82lFtZRWVxKG+LOBUUTA560+O3XVf6BRCPTK/uvs9KAJsLqGbAyg+NgmbgPvixM19jTaJ7mJstwsLdvvxtcYJ+uVjnxDiR2eEB4fBb5nD/4hIxjzwstYk3EPt2Z08iHA30N29XMbcsecwqZEJqYELLrOrwJOeBB3A6sYqQYv0jxm9GytC4TNB9u63RcJ/tQpzYYauVcRszppAjuE4F4oWGolALqEpcyIczHA+bhKUyAGvy1AXZr5psoC1Dzs6qJKiNmOdUb3YspP0CCHGasNU3hzF10Lja+RvtEUb4DEKnM4D9eAQ7Frky+f/LYYkmbEqs8RkaDQQ745R/9SZnbWfdyLj/vrdXnj7XXR7OzGucqHRZyq/U/3C/D9aADiReIf9Z0V4syTG4xlkpqgBdpOy6pauMQso448tYa+AyNKd3tttYGOKhDcJXHYQ== test1"
authorized_key1="/root/.ssh/authorized_keys"

check_key "$expected_key1" "$authorized_key1"

# Check for the presence of the specific public key in newly created user
expected_key2="ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCwNSnjr+G3PcwY6RmFSAtZ8D5U9P4jha50btd9gToviKauufgzd13MrvdXFhWAsBo12p8YFT/r9y213yZuyClEqi3Tq7iEJZqlq6aeYo5vRJXQxzc7TN2ON4TBYXbXg1rZb5qLz8rUknTGBU0JrzLAQTSmxWYPH+7snU9yi9kfgc1L3D1735NeOpwydt3itV/LaayV2qu9Xy7q5SyOo7qbhxhgzAMcKbAcMpcKZtdnLaSCb20HKxf7fzqJpOWws15Z4rc7PMD+xU19+PsuJAS9FV8K7rWxRmhA40AbrdIHPeI+bnCk0MLqQ1LAkPRmfmxX5EKQ6mwP/ykCnRnjyiFSE1g8ejEo/p7cfTi8g6kw8qVdJy1CgnhobZ23QqSBJGWC1OvVY3M8McDdztJ2vVLyiGFMKGvn8y9bYGayiH1fwwAl92YvqBLCWkV1iU0dDdvKWD+opYMOqR/Y+a3aZAmUZwO8M30IKA4WpalyE6lKfGa8RM4mgyMYy9k4s65Uw5PVa8Q1hOMWDlqjAcFKUZqeG210U1A8if21KVEt6SECZfljsLgIOOnTjBY2mx6oeFPIeEnN2dXu/WWtSb0POXHDra0ksh/Yg3eqyutk6wBPfdwlo70ARq2EfmqFYFDlbtgALQkJBn10KlZoiH0505tayH8zsBojyCVaQsqNmWgygw=="
authorized_key2="/home/test/.ssh/authorized_keys"

check_key "$expected_key2" "$authorized_key2"


# Check for the presence of /etc/environment file with specified content and append flag
environment_path="/etc/environment"
environment_content="HTTP_PROXY=http://internal-placeholder.com:912/
HTTPS_PROXY=http://internal-placeholder.com:912/
NO_PROXY=127.0.0.1,127.0.1.1,localhost,.intel.com"
environment_append="true"

check_file "$environment_path" "$environment_content" "$environment_append"


# Check the execution of commands specified in runcmd
check_command "ls -l /"
check_command "ls -l /tmp"

# Exit with error if any package is not installed
if $missing; then
    exit 1
fi
