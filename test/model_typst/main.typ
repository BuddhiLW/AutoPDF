#set page(paper: "a4", margin: 2cm)
#set text(size: 11pt)
#let pill(body) = box(inset: 3pt, radius: 2pt, stroke: 0.5pt, body)

= delim[[ escapeTypst .title ]]

Author: delim[[ escapeTypst .author ]] #h(1fr) delim[[ escapeTypst .date ]]

delim[[ escapeTypst .summary ]]

== Items

delim[[ range .items -]]
- #pill[delim[[ escapeTypst .name ]]]: delim[[ escapeTypst .note ]]
delim[[ end ]]

Typst brackets outside actions are untouched: #pill[a[b]].
