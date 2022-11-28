set -e
set -x

substate_cli=/home/xihu5895/record-replay/cmd/substate-cli/substate-cli 
step=1000000
substate_dir=/mnt/backup/substates/
out_folder=/mnt/backup/dep_trace/
workers=50

# mkdir -p $out_folder/9-10M
# $substate_cli dependency-trace --workers=$workers --substatedir=$substate_dir/substate.ethereum.0-10M  --output-dir=$out_folder/9-10M 9000000 10000000
 for i in {10..13};
 do
     start=$i
     end=$(($i+1))
     start_range=$(($start*$step))
     end_range=$(($start_range+$step))
     mkdir -p $out_folder/$start-${end}M
     $substate_cli dependency-trace --workers=$workers --substatedir=$substate_dir/substate.ethereum.$start-${end}M --output-dir=$out_folder/$start-${end}M $start_range $end_range
 done
