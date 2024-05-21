Simple script to extract and parse PDF text only with pure Go.
The script make use of the PDFcpu library to open and validate a pdf file. Then it makes use of rsc.io/pdf library to extract the single text cells in the pdf. Finnaly the text cells are parsed based on the position, font and size.
This is a very raw script, but is the one working among all the other free alternatives. Feel free to contribute.

In order to use it just sobstitute the inFile variable in the main() with the path of the desired pdf.
