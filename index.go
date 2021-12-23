package main

const CodexOutputTemplate = `
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" lang="" xml:lang="">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=yes"/>
  <title>codex</title>

  <link rel="icon" type="image/svg" href="static/codex.svg"/>

  <script src="https://code.jquery.com/jquery-3.6.0.min.js"> </script>
  <script src="https://unpkg.com/lunr@2.3.9/lunr.js"></script>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/mark.js/8.11.1/jquery.mark.min.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-chtml-full.js" type="text/javascript"></script>

  <link href="https://fonts.googleapis.com/css2?family=Inter&family=Ubuntu+Mono&display=swap" rel="stylesheet">

  <script src="static/codex.js"></script>
  <link rel="stylesheet" href="static/pandoc.css"/>
  <link rel="stylesheet" href="static/codex.css"/>
</head>

<body>
  <nav>
	<div id="search">
		<input type="text" placeholder="search" autocomplete=off name="search-input" size="20">
		<label for="search-input"> <!-- used by codex.js --> </label>
	</div>

	<div id="files">
		<!-- codex file navigation -->
	</div>
  </nav>

  <main>
    <!-- codex contents -->
  </main>

  <div id="full-screen-modal" class="inactive"> <!-- used by codex.js --> </div>
</body>

</html>`
