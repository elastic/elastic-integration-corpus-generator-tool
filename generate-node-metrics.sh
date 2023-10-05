function exit_handler()
{
    echo $1 $2 $3 $4
    replace "hour" $1 "placeholder_hour"
    replace "minute" $2 "placeholder_min"
    replace "second" $3 "placeholder_sec"
    replace "timedate" $4 "placeholder_date"
    rm data.json
    echo "exiting"
    exit
}

function replace()
{
    tmpfile=$(mktemp)
    cp ./assets/templates/kubernetes.node/configs.yml "$tmpfile" 
    awk '/name: '$1'/{ rl = NR + 1 } NR == rl { gsub( "'$2'","'"$3"'") } 1' "$tmpfile" > ./assets/templates/kubernetes.node/configs.yml
    rm "$tmpfile" 
}

Help()
{
   # Display Help
   echo "Generate real time kubernetes node data and send them to Elasticsearch"
   echo
   echo "Prerequisites:"
   echo "Set ELASTIC_USERNAME, ELASTIC_PASS and ELASTIC_HOST environment variables"
   echo "Example: export ELASTIC_USERNAME=elastic && export ELASTIC_PASS=changeme && export ELASTIC_HOST=http://localhost:9200"
   echo
   echo "Syntax: bash generate-node-metrics.sh NUMBER_OF_EVENTS_PER_10_SECONDS [-h]"
   echo "Example: bash generate-node-metrics.sh 1000"
   echo "options:"
   echo "h     Print Help."
   echo
}

# Get the options
while getopts ":h" option; do
   case $option in
      h) # display Help
         Help
         exit;;
   esac
done

if [ -z "${ELASTIC_USERNAME}" ]; then
   echo "Variable ELASTIC_USERNAME isn't set. Exiting"
   exit
fi
if [ -z "${ELASTIC_PASS}" ]; then
   echo "Variable ELASTIC_PASS isn't set. Exiting"
   exit
fi
if [ -z "${ELASTIC_HOST}" ]; then
   echo "Variable ELASTIC_HOST isn't set. Exiting"
   exit
fi

events=$1
three=3
timedate=`date +%Y-%m-%d`
hour=`date +"%H"`
hour=10#$hour
utchour="$((hour-three))"
if [ $utchour -lt 10 ]
then
    utchour=0$utchour
fi
minute=`date +"%M"` 
second=`date +"%S"`.328Z
oldutchour=$utchour
oldminute=$minute
oldsecond=$second

replace "timedate" "placeholder_date" $timedate
replace "hour" "placeholder_hour" $utchour
replace "minute" "placeholder_min" $minute
replace "second" "placeholder_sec" $second

trap "exit_handler $oldutchour $oldminute $oldsecond $timedate" EXIT
while true
do
    sleep 10
    hour=`date +"%H"`
    hour=10#$hour
    utchour="$((hour-three))"
    if [ $utchour -lt 10 ]
    then
        utchour=0$utchour
    fi
    minute=`date +"%M"` 
    second=`date +"%S"`.328Z
    replace "hour" $oldutchour $utchour
    replace "minute" $oldminute $minute
    replace "second" $oldsecond $second
   
    oldutchour=$utchour
    oldminute=$minute
    oldsecond=$second
    trap "exit_handler $oldutchour $oldminute $oldsecond $timedate" EXIT
    output=`./elastic-integration-corpus-generator-tool generate-with-template ./assets/templates/kubernetes.node/gotext.tpl ./assets/templates/kubernetes.node/fields.yml -c ./assets/templates/kubernetes.node/configs.yml -y gotext -t $1`
    #Define multi-character delimiter
    delimiter=" "
    #Concatenate the delimiter with the main string
    string=$output$delimiter

    #Split the text based on the delimiter
    myarray=()
    while [[ $string ]]; do
    myarray+=( "${string%%"$delimiter"*}" )
    string=${string#*"$delimiter"}
    done

    generatedfile="`echo ${myarray[2]} ${myarray[3]}`"
    echo ""
    echo $generatedfile
    echo $utchour:$minute:$second
    mv "$generatedfile" data.json
    curl --location --request POST -u ${ELASTIC_USERNAME}:${ELASTIC_PASS} ${ELASTIC_HOST}/metrics-kubernetes.node-default/_bulk  --header 'Content-Type: application/x-ndjson' --data-binary @data.json
done