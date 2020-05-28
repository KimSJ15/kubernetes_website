#!/bin/bash

# set K8S_WEBSITE in your env to your docs website root
# Note: website/content/<lang>/docs
CONTENT_DIR=${K8S_WEBSITE}/content

# 16 langs
# de en es fr hi id it ja ko no pl pt ru uk vi zh

declare -a DIRS=("concepts" "contribute" "home" "reference" "setup" "tasks" "tutorials")
declare -a EMPTY_STMTS=("body" "discussion" "lessoncontent" "overview" "steps")
declare -a REPLACE_STMTS=("cleanup" "objectives" "options" "prerequisites" "seealso" "synopsis" "whatsnext")
declare -a CONTENT_TYPES=("concept" "task" "tutorial" "tool_reference")
END_CAPTURE="{{% \/capture %}}"
CONTENT_TEMPLATE="content_template:"

# replace or remove capture statements
function replace_capture_stmts {
  echo "i:""$i"
  if [ -d "$1" ] ; then
    for i in `ls $1`; do
    replace_capture_stmts "${1}/${i}"
    done
  else
    if [ -f "$1" ] ; then
      ls -f $1 | while read -r file; do
        for stmt in "${EMPTY_STMTS[@]}" ; do
          CAPTURE_STMT="{{% capture ""$stmt"" %}}"
          COMMENT_REPLACE="<!-- ""$stmt"" -->"
          sed -i -e "s/${CAPTURE_STMT}/${COMMENT_REPLACE}/g" $1
        done

        for stmt in "${REPLACE_STMTS[@]}" ; do
          CAPTURE_STMT="{{% capture ""$stmt"" %}}"
          HEADING_STMT="## {{% heading \"""$stmt""\" %}}\n"
          echo "HEADING STMT TO ADD:""$HEADING_STMT"
          sed -i -e "s/${CAPTURE_STMT}/${HEADING_STMT}/g" $1
        done

        sed -i -e "s/${END_CAPTURE}//g" $1

        # replace content_template: templates/<template_type> with
        # content_template: <type>
        #sed -i -e "s/^${CONTENT_TEMPLATE}/# ${CONTENT_TEMPLATE}/g" $1
        for t in "${CONTENT_TYPES[@]}" ; do
          sed -i -e "s/content_template:[[:space:]]*templates\/$t/content_type: $t/g" $1
        done
      done
    else
      exit 1
    fi
  fi
}

# change to docs content dir
cd $CONTENT_DIR

for langdir in `ls $CONTENT_DIR`; do
  # Testing with a couple of langs to start
  if [ $langdir = "en" ] ; then
  LANGDIR="$CONTENT_DIR""/""$langdir""/docs"

  for d in "${DIRS[@]}"; do
    ROOTDIR="${LANGDIR}""/""$d"
    cd ${ROOTDIR}
    for i in `ls ${ROOTDIR}`; do
      replace_capture_stmts "${ROOTDIR}""/""$i"
    done
  done
  fi
done
