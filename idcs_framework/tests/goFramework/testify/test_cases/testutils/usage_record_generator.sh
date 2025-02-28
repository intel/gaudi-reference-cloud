#!/bin/sh

## CONSTANTS
CREATE_API='proto.MeteringService/Create'
HOST=localhost
PORT=8080

# -------------------------------------------------------------------- #

## Create records 
create_records() 
{    
    c_ca=0
    time_counter=1

    ## Calculating records per resource per cloud_Account
    TOTAL_TX=$(( $TOTAL_RECORDS / ($TOTAL_CLOUD_ACCOUNTS * $TOTAL_RESOURCES) ))

    while [ $c_ca -lt $TOTAL_CLOUD_ACCOUNTS ]
    do
        ## Random cloud_accound_id
        ca=$(shuf -i200-1000 -n1)

        c_resources=0
        while [ $c_resources -lt $TOTAL_RESOURCES ]
        do
            ## Generate resource_id
            rid=$(uuidgen)

            c_ta=0  
            while [ $c_ta -lt $TOTAL_TX ]
            do
                ## Timestamp calculation for each record = current_time + step in mins
                ## step would increase by $INTERVAL_MINS for each record timestamp
                step=`expr $time_counter \* $INTERVAL_MINS`

                dt=$(date -d '+'$step' minutes'  '+%FT%T.000Z')

                ## New transaction id for each record
                txid=$(uuidgen)

                ## Calling create api to create records in the database.
                grpcurl -d '{"transactionId": "'$txid'",
                 "resourceId": "'$rid'", "cloudAccountId": "'$ca'",
                 "timestamp": "'$dt'",
                 "properties":{"instance":"small"} }' -plaintext $HOST:$PORT $CREATE_API >/dev/null

                ## Validating output 
                if [ $? -eq 0 ]; then
                    echo "INFO: Successfully created record for cloudAccountId:$ca, resourceId:$rid, time $dt "
                else
                    echo "ERROR: Failed to create record. Exiting ..."; exit 1
                fi
                time_counter=`expr $time_counter + 1`
                c_ta=`expr $c_ta + 1`
            done
            c_resources=`expr $c_resources + 1`
        done
        c_ca=`expr $c_ca + 1`
    done

}

# -------------------------------------------------------------------- #

help()
{
    echo "Usage: ./usage_record_generator.sh [ --host ] 
               [ -r | --records ]
               [ -nr | --num_resources ]
               [ -i | --interval_mins ]
               [ -ca | --cloud_accounts ]
               [ -h | --help ]\n
        Ex: ./usage_record_generator.sh --host localhost
                     --records 20 --num_resources 5 
                     --interval_mins 5 --cloud_accounts 2         
        Note: 
        1. Make sure the number of records specified are multiple of (cloud_accounts * num_resources).
        2. Metering-api server should be up & running on the host specified by --host parameter"
    exit 2
}

# -------------------------------------------------------------------- #

## Parse options
while [ $# -gt 0 ]; do
    case "$1" in
    --host)
        HOST="$2"
        shift 
        ;;
    -r|--records)
        TOTAL_RECORDS="$2"
        shift 
        ;;
    -nr|--num_resources)
        TOTAL_RESOURCES="$2"
        shift 
        ;;
    -i|--interval_mins)
        INTERVAL_MINS="$2"
        shift 
        ;;
    -ca|--cloud_accounts)
        TOTAL_CLOUD_ACCOUNTS="$2"
        shift 
        ;;
    -h|--help)
        help
        shift 
        ;;
    *) echo "ERROR: Invalid input"; help; exit 1;;
    esac
    if [ $# -lt 1 ]; then
        echo "ERROR: Unexpectedly ran out of arguments"; exit 1;
    fi
    shift 
done
create_records
echo "INFO: Script exiting..."
