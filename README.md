# rdm-emailer
Sending emails to DR system users via SMTP.

## Requirement

The [GO](https://golang.org) compiler is required to build the source code.

## Build

```bash
$ make
```

The executables will be created in the `bin/` directory.

## Usage

Firstly define an email template, an example can be found in the [template.txt](template.txt) file.  Secondly make a recipient list as shown in the example file [recipients.csv](recipients.csv).

Run the following command to send the email to all recipients:

```bash
$ ./bin/sendmail -f from_address -l recipients.txt -n smtp_host -p smtp_port -u smtp_username -s smtp_password template.txt
```

The command-line options are provided with `-h` option:

```bash

Usage: ./bin/sendmail [OPTIONS] <template>

OPTIONS:
  -f email
    	set the sender's email address. (default "no-reply@donders.ru.nl")
  -l path
    	set path of the file containing a list of recipients.
  -n hostname
    	set the network hostname of the SMTP server. (default "smtp-auth.ru.nl")
  -p port
    	set the network port of the SMTP server. (default 25)
  -s password
    	set SMTP password for PLAIN authentication.
  -u username
    	set SMTP username for PLAIN authentication.
```
