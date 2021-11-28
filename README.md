# Codex

Codex turns an unstructured pile of heterogeneous documents into a single
interactive web document.

Your input documents maybe in markdown, TeX, reStructuredText, docx, or any
other format supported by [pandoc].

[pandoc]: https://pandoc.org/

## Usage

```
$ codex notes/*
$ codex nodes/* --outdir somedir/ # override default output directory
$ codex nodes/* --ws              # watch and serve
```

Tested with Go 1.17.

## How does it work?

Codex has 4 pieces:

1. **Parse**: Codex accepts a wide range of input formats, thanks to [pandoc].
   The output of the parsing step is a single HTML tree containing all input
   documents in their original order.
2. **Transform**: this is where the core idea is implemented. Given the DOM tree
   of the previous step, Codex traverses and transforms the tree in such a way
   to make it match its own [semantic structure](#semantic-trees).
3. **Server**: optionally, you can ask Codex to "watch and serve" your input
   files. In this mode, the output document is served over HTTP, and the server
   watches input files for changes, submitting incremental changes to the UI
   over a websocket.
4. **Client**: the interactive UI turns the output markup into a searchable,
   easy to navigate document.

### Semantic Trees

Consider this document:
```html
 <h1> Title </h1>
 <p> Paragraph </p>
 <h3> Section </h3>
 <p> Section paragraph </p>
 <h2> Chapter </h2>
 <p> Chapter paragraph </p>
 ```

This DOM tree is a flat list of siblings while its semantic structure is quite
different:
```html
<h1> Title </h1>
  <p> Paragraph </p>
  <h3> Section </h3>
    <p> Section paragraph </p>
    <h2> Chapter </h2>
      <p> Chapter paragraph </p>
```

Codex transforms the DOM in such a way that it matches its own semantic
structure. The building block of this is a **node**:

```html
<node depth=2>        <!-- eg <div class="node node-depth-2 ..."> ... -->
  <node-head>
    <h4> ... </h4>
  </node-head>
  <node-body>
    ...               <!-- recurse, child nodes live here -->
  </node-body>
</node>
```

**Are semantic trees well-defined?**

A DOM tree for which it's clear what this "semantic tree" is must follow rigid
rules. For example, if you wrap an `<h1>` in 5 different `<div>`s, how does the
"headness" of `<h1>` propagate up its ancestry?

Luckily, the kinds of trees that can possibly come out of a typical, say,
markdown or docx file, *are* indeed quite limited; written documents can only
represent trees that can be flattened into a linear order.

The main technical assumption that Codex makes about its input DOM tree is this:

**All headings in the document are immediate children of `<body>`**

With this, we can then define clearly what the semantic tree is, how its related
to the DOM tree, and how to transform the DOM into its semantic structure, see
`treeify.go`.

### Releative Depths

Codex node depth calculation is file-scoped and relative to context.

**Relative to context**: its not the heading types that dictate node depths,
but their relationship with their neighbors. For example, the following two
markdown files produce the same tree:

```md
# H1

# H1

## H2
```
```md
## H2

## H2

##### H5


**File-scoped**: each file is reasoned about on its own. This is consistent
with the above rule and implies no tree interference between different input
files. For example, the two example files above, if concatenated by Codex,
will produce a tree consisting of the same subtree twice, one for each of the
input files.
