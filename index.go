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
  <link href="https://fonts.googleapis.com/css2?family=Ubuntu+Mono&amp;display=swap" rel="stylesheet"/>

  <script src="static/codex.js"></script>
  <link rel="stylesheet" href="static/pandoc.css"/>
  <link rel="stylesheet" href="static/codex.css"/>
</head>

<body>
  <div id="top-bar">
    <input id="search-input" type="text" placeholder="search" autocomplete=off name="search-input" size="20">
    <label for="search-input"> <!-- used by codex.js --> </label>
  </div>

  <main>
    <!-- codex contents -->
  </main>

  <div id="full-screen-modal" class="inactive"> <!-- used by codex.js --> </div>
</body>

</html>`
