(TeX-add-style-hook
 "main"
 (lambda ()
   (TeX-add-to-alist 'LaTeX-provided-package-options
                     '(("inputenc" "utf8") ("fontenc" "T1") ("xcolor" "dvipsnames") ("pgfornament" "object=vectorian")))
   (TeX-run-style-hooks
    "latex2e"
    "scrartcl"
    "scrartcl10"
    "inputenc"
    "fontenc"
    "xcolor"
    "pgfornament")
   (LaTeX-add-xcolor-definecolors
    "fondpaille"))
 :latex)

