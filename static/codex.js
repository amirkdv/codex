const debounce = (wait, func) => {
  let timeout;

  return (...args) => {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };

    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
};

class Codex {
  constructor(root) {
    this.addFullScreenButtons();
    this.initSearch();

    this.initNav();
    this.initHighlighting();
    this.initFolding();
    this.initFullScreen();
    this.initWebSocket();
  }

  addFullScreenButtons() {
    $('.node:not(:has(.node))').addClass('node-leaf')
    $('.node').each((idx, elem) => {
      $(elem).append(`<div class="full-screen-button"> ⤢ </div>`);
    });
  }

  initSearch() {
    delete this.searchIndex
    this.searchIndex = lunr(config => {
      config.ref('id');
      config.field('text');

      $('.node-leaf').each((i, elem) => {
        config.add({id: elem.id, text: elem.innerText});
      });
    })

    $('#search input').on('keyup', debounce(400, event => {
      if (event.target.value == '') {
        $('.node').removeClass('d-none')
        $('#search label').text('');
        return;
      }
      $('.node').addClass('d-none');
      // query syntax: https://lunrjs.com/guides/searching.html
      // bug: colon is broken because it gets interpreted as "field query"
      const hits = this.searchIndex.search(event.target.value);
      $('body').unmark({
        done: () => {
          for (const hit of hits) {
            $(`#${hit.ref}`).removeClass('d-none');
            $(`#${hit.ref}`).parents('.node').removeClass('d-none');
            $(`#${hit.ref}`).mark(Object.keys(hit.matchData.metadata));
          }
        }
      });

      $('label[for="search-input"]').text(hits.length ? `${hits.length} nodes` : 'no matches');
    }));
  }

  initNav() {
    $('main article[codex-source]').each((idx, elem) => {
      const $article = $(elem);
      const fname = $article.attr('codex-source');
      $('nav #files').append(`
        <div class="nav-file" codex-source="${fname}">
          <div class="file-name"> ${fname} </div>
          <div class="last-updated"> <!-- popualted later --> </div>
        </div>
      `);
      this.renderLastUpdated($article);
    });

    $('main').on('mouseenter', '.node', event => {
      const $article = $(event.target).closest('article[codex-source]');
      this.navForArticle($article).find('.file-name').addClass('bold');
    });

    $('main').on('mouseleave', '.node', event => {
      const $article = $(event.target).closest('article[codex-source]');
      this.navForArticle($article).find('.file-name').removeClass('bold');
    });

    $('.nav-file').on('click', event => {
      const fname = $(event.target).closest('.nav-file').attr('codex-source');
      $(`article[codex-source="${fname}"]`)[0].scrollIntoView();
    });
  }

  navForArticle($article) {
    const fname = $article.attr('codex-source');
    return $(`#files .nav-file[codex-source="${fname}"]`)
  }

  renderLastUpdated($article) {
    const fname = $article.attr('codex-source');
    const mtime = (new Date($article.attr('codex-mtime'))).toLocaleString();
    $(`nav #files div[codex-source="${fname}"] .last-updated`).html(mtime);
  }

  initHighlighting() {
    // nuance: moving cursor through nodes, in and out of nodes within a node
    // this needs to be a single event handler for the entire DOM
    $('body').on('mousemove mouseenter', '.node', event => {
      $('.node').removeClass('highlight');
      const $node = $(event.target);
      $node.addClass('highlight');
      $node.parents('.node').addClass('highlight')
    });
    $('body').on('mouseleave', '.node', event => {
      const $node = $(event.target);
      $node.children('.node').removeClass('highlight');
      $node.removeClass('highlight');
    });
  }

  initFolding() {
    $('main').on('click', '.node-head', event => {
      if (event.target.tagName == 'A') {
        return;
      }
      $(event.target).closest('.node').toggleClass('collapsed');
    });
  }

  initFullScreen() {
    // clicking on the full screen button populates the modal with the current
    // node and blurs the rest into the background
    $('body').on('click', '.full-screen-button', event => {
      $('.node').removeClass('highlight');
      $('body').addClass('full-screen');
      $('#full-screen-modal').empty();
      $('#full-screen-modal').append($(event.target).closest('.node')[0].innerHTML);
      $('#full-screen-modal').removeClass('inactive');
    });

    // While the modal is active, a click outside of it removes it
    $('html').on('click', event => {
      if (!this.isInFullScreen()) {
        return;
      }
      if ($(event.target).closest('.full-screen-button').length) {
        // this was the click that initiates full screen
        return;
      }
      if ($(event.target).closest('#full-screen-modal').length) {
        // this was a click inside the full screen modal
        return;
      }
      this.exitFullScreen();
    });

    // While the modal is active, pressing Escape exits the full screen modal
    $(document).keyup(event => {
      if (this.isInFullScreen() && event.key === 'Escape') {
        this.exitFullScreen();
      }
    });
  }

  isInFullScreen() {
    return $('body').hasClass('full-screen');
  }

  exitFullScreen() {
    $('body').removeClass('full-screen');
    $('#full-screen-modal').addClass('inactive');
  }

  initWebSocket() {
    this.websocket = new WebSocket(`ws://${document.location.host}/ws`);
    this.websocket.onmessage = async (msg) => {
      // assumes msg is html for an <article>, note: article ~ input doc
      const data = await msg.data;
      const text = (typeof data === 'string') ? data : await data.text();
      this.onServerUpdate(text);
    }
  }

  onServerUpdate(html) {
    const parser = new DOMParser();
    const $newDoc = $(parser.parseFromString(html, 'text/html'));
    const $article = $newDoc.find('article');

    const codexSource = $article.attr('codex-source');
    $(`main article[codex-source="${codexSource}"]`).replaceWith($article);

    this.addFullScreenButtons();
    this.initSearch();
    this.renderLastUpdated($article);

    // tell MathJax to look for unprocessed math and typeset it
    MathJax.typeset();
  }
}

$(document).ready(() => {
  new Codex();
});
