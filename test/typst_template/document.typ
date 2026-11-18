#set page(paper: "a4")
#let f(body) = [*#body*]

= delim[[ .title ]]

Author: delim[[ escapeTypst .author ]]

#f[a[b]]

Closing brackets in text stay put: ]] and #f[x[y[z]]].
