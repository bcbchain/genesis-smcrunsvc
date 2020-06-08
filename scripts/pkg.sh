#!/bin/bash

function GetDirs() {
  IDIRS=()
  i=1
  for _ in $(cat ./scripts/.distignore)
  do
    NUM=$i
    IDIR=$(awk 'NR=='$NUM' {print $1}' ./scripts/.distignore)
    if [[ -n "$IDIR" ]]; then
      IDIRS[$i]=$IDIR
    fi

    : $(( i++ ))
  done

  for f in `ls -l $PWD`
  do
    if [[ -d "$f" ]];then
      b=0
      for id in "${IDIRS[@]}"
      do
        if [[ "$id" == "$f" ]];then
          b=1
        fi
      done

      if [[ $b == 0 ]];then
        TDIRS[${#TDIRS[*]}]=$f
      fi
    fi
  done
  return 0
}

DIST_DIR=./build/dist/
TDIRS=()
cd ..

echo "==> Removing old directory..."
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

genesisDir="./genesis/"
#smcRunSvr="genesis-smcrunsvc_"

echo "==> Tar genesis files..."
GetDirs

mkdir -p temp

UNAME=`uname`

DARWIN="Darwin"

for d in "${TDIRS[@]}"
do
  if [[ "$d" == "genesis" ]];then
    mkdir -p temp/genesis/src
    cp -r "$d"/* temp/genesis/src

    if [[ $UNAME == $DARWIN ]];then
      sed -i "" "10s/genesis-smcrunsvc\/genesis\/genesis/genesis/" temp/genesis/src/cmd/smcrunsvc/smcrunsvc.go
      sed -i "" "6s/genesis-smcrunsvc\/genesis\/stubcommon/stubcommon/" temp/genesis/src/genesis/stub_interface.go
      sed -i "" "7s/genesis-smcrunsvc\/genesis\/stubcommon/stubcommon/" temp/genesis/src/genesis/stub_interface.go
      sed -i "" "4s/genesis-smcrunsvc\/genesis\/stubcommon/stubcommon/" temp/genesis/src/genesis/stub_method.go
      sed -i "" "5s/genesis-smcrunsvc\/genesis\/stubcommon/stubcommon/" temp/genesis/src/genesis/stub_method.go
    else
       sed -i "10s/genesis-smcrunsvc\/genesis\/genesis/genesis/" temp/genesis/src/cmd/smcrunsvc/smcrunsvc.go
       sed -i "6s/genesis-smcrunsvc\/genesis\/stubcommon/stubcommon/" temp/genesis/src/genesis/stub_interface.go
       sed -i "7s/genesis-smcrunsvc\/genesis\/stubcommon/stubcommon/" temp/genesis/src/genesis/stub_interface.go
       sed -i "4s/genesis-smcrunsvc\/genesis\/stubcommon/stubcommon/" temp/genesis/src/genesis/stub_method.go
       sed -i "5s/genesis-smcrunsvc\/genesis\/stubcommon/stubcommon/" temp/genesis/src/genesis/stub_method.go
    fi
  else
    cp -r "$d" temp/
  fi
done

cd temp
tar -zcf "../$DIST_DIR$project_name""_$VERSION".tar.gz genesis
cd ..
rm -rf temp "$genesisDir""temp"

# Make the checksums.
pushd "$DIST_DIR" > /dev/null
shasum -a256 ./* > "$project_name"_SHA256SUMS
popd >/dev/null

echo ""
echo "==> Build results:"
echo "==> Path: "../../$DIST_DIR""
echo "==> Files: "
ls -hl "$DIST_DIR"