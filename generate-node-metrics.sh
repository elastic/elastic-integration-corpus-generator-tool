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
while true
do
    datenow=`date -u +"%FT%T.%3Z"`
    echo $datenow
    output=`./elastic-integration-corpus-generator-tool generate-with-template ./assets/templates/kubernetes.node/schema-b/gotext.tpl ./assets/templates/kubernetes.node/schema-b/fields.yml -c ./assets/templates/kubernetes.node/schema-b/configs.yml -y gotext -t $1 -n $datenow`
    
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
    mv "$generatedfile" data.json
    curl --location --request POST -u ${ELASTIC_USERNAME}:${ELASTIC_PASS} ${ELASTIC_HOST}/metrics-kubernetes.node-default/_bulk  --header 'Content-Type: application/x-ndjson' --data-binary @data.json
    rm data.json
    sleep 10
done