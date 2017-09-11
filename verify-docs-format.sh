#!/bin/bash

no_entry=false
no_entry_counter=0
no_title=false
no_title_counter=0

# Verify all docs/.../*.md files 
# Skip checking autogenerated files in some folders
# (docs/api-reference/v1.5, docs/user-guide/kubectl/v1.5, and
# docs/resources-reference/v1.5)
for file in `find docs -name "*.md" -type f`; do 
  # Skip checking all files in the following folders 
  if [[ "${file}" == "docs/api-reference/v1."* ]] ||
     [[ "${file}" == "docs/user-guide/kubectl/v1."* ]] ||
     [[ "${file}" == "docs/resources-reference/v1."* ]]; then
    continue
  fi 

  # 1. TOC check:
  #    Check they are referenced in at least one of _data/*.yml files.
  #    Skip checking files in skip_toc_check.txt
  if ! grep -q "${file}" skip_toc_check.txt; then
    path=${file%.*}
    # abc/index.md should point to abc, not abc/index
    path=${path%%index}
    if ! grep -q "${path}" _data/*.yml; then
      echo "Error: ${file} doesn't have an entry in the table of contents under _data/*.yml" 
      no_entry=true
      no_entry_counter=$[no_entry_counter+1]
    fi
  fi

  # 2. Title check:
  #    Check they have a proper title.
  #    Skip checking files in skip_title_check.txt.
  #    Title should start with "title:" and can have several spaces/tabs between
  #    non-space/tab content. They should also be inside the markdown header.
  #    For example, "title:", " title: abc", and "title:" aren't valid, 
  #    but "title: abc", "title:def" and "title:    def ghi" are both valid.
  if [[ "${file}" == "docs/user-guide/kubectl/kubectl"* ]]; then 
    # Skip checking auto-generated kubectl docs since its first heading matches title 
    continue
  fi 
  if ! grep -q "${file}" skip_title_check.txt; then
    if ! grep -q "^title:\s*[^\s]" ${file}; then 
      echo "Error: ${file} doesn't have a proper title defined!"
      no_title=true
      no_title_counter=$[no_title_counter+1]
    fi
  fi
done

if ${no_entry}; then 
  echo "Found ${no_entry_counter} files without entries. For how to fix it, see http://kubernetes.io/docs/home/contribute/write-new-topic/#creating-an-entry-in-the-table-of-contents"
  exit 1
fi

if ${no_title}; then 
  echo "Found ${no_title_counter} files without titles."
  exit 1
fi
