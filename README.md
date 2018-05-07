# rdm-emailer
Sending bulk emails to DR system users via SMTP.

## Requirement

The [GO](https://golang.org) compiler is required to build the source code.

## Build

```bash
$ make
```

The executables will be created in the `bin/` directory.

## Usage

Firstly define an email template, an example can be found in the [template.txt](template.txt) file.  Secondly make a recipient list as shown in the example file [users.txt](users.txt).

Run the following command to send the email to all recipients:

```bash
$ ./bin/sendmail -f from_address -l users.txt -n smtp_host -p smtp_port -u smtp_username -s smtp_password template.txt
```

The command-line options are provided with `-h` option:

```bash

Usage: ./bin/sendmail [OPTIONS] <emailTemplate>

OPTIONS:
  -f string
    	set the sender's email address. (default "no-reply@donders.ru.nl")
  -l string
    	set path of file containing list of user emails and names.
  -n string
    	set the network hostname of the SMTP server. (default "smtp-auth.ru.nl")
  -p int
    	set the network port of the SMTP server. (default 25)
  -s string
    	set SMTP password for PLAIN authentication.
  -u string
    	set SMTP username for PLAIN authentication.
```
