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
    // TODO this needs to be repeated everytime a new node gets through the door
    // each node gets a full screen button
    $('.node').each((idx, elem) => {
      $(elem).append(`<div class="full-screen-button"> ⤢ </div>`);
    });

    this.initSearch();
    this.initFileLabels();
    this.initHighlighting();
    this.initHeadToggle();
    this.initFullScreen();
  }

  initSearch() {
    this.searchIndex = lunr(config => {
      config.ref('id');
      config.field('text');

      // TODO revisit the fact that the text of a node is indexed as part of
      // itself and all its parents!
      // TODO this needs to happen on websocket updates too
      $('.node-leaf').each((i, elem) => {
        config.add({id: elem.id, text: elem.innerText});
      });
    })

    $('#search-input').on('keyup', debounce(400, event => {
      // TODO refactor
      if (event.target.value == '') {
        $('.node').removeClass('d-none')
        $('label[for="search-input"]').text('');
        return;
      }
      $('.node').addClass('d-none');
      // query syntax: https://lunrjs.com/guides/searching.html
      // bug: colon is broken because it gets interpreted as "field query"
      const hits = this.searchIndex.search(event.target.value);
      for (const hit of hits) {
        $(`#${hit.ref}`).removeClass('d-none');
        $(`#${hit.ref}`).parents('.node').removeClass('d-none');
      }
      $('label[for="search-input"]').text(hits.length ? `${hits.length} nodes` : 'no matches');
    }));
  }

  initFileLabels() {
    $('.node').each((idx, elem) => {
      const $elem = $(elem);
      const fname = $elem.attr('codex-source');
      const mtime = (new Date($elem.attr('codex-mtime'))).toLocaleString();
      const $label = $(`<div class="source-file-label"> ${fname} (last updated: ${mtime}) </div>`);
      $elem.children('.node-head').prepend($label);
    });
  }

  initHighlighting() {
    // nuance: moving cursor through nodes, in and out of nodes within a node
    // this needs to be a single event handler for the entire DOM
    $('body').on('mousemove mouseenter', '.node', event => {
      $('.node').removeClass('highlight');
      $(event.target).addClass('highlight');
      $(event.target).parents('.node').addClass('highlight')
    });
    $('body').on('mouseleave', '.node', event => {
      $(event.target).children('.node').removeClass('highlight');
      $(event.target).removeClass('highlight');
    });
  }

  initHeadToggle() {
    $('body').on('click', '.node-head', event => {
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
}

$(document).ready(() => {
  new Codex();
});
