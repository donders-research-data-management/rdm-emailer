docker image pull docker.isc.ru.nl/rdr/tool/rdr-emailer:latest
docker run -v `pwd`/recipients.csv:/recipients.csv \
             -v `pwd`/template.txt:/template.txt \
             rdr-emailer /rdr-emailer \
             -l /recipients.csv \
             -n smtp-bulk.ru.nl -p 25 \
             /template.txt
#             -f from_address
#             -u smtp_username -s smtp_password
