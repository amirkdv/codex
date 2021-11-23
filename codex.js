const debounce = (wait, func) => {
  let timeout;

  // FIXME arrow?
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };

    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
};

function bootSearch() {

  // TODO experiment with fuse.js is that better?
  // https://fusejs.io/demo.html

  const idx = lunr(function () {
    this.ref('id');
    this.field('text');

    // TODO revisit the fact that the text of a node is indexed as part of
    // itself and all its parents! There's some old code that deals with this.
    $('.node').each((i, elem) => {
      this.add({id: elem.id, text: $(elem).text()});
    });
  })

  $('#search-input').on('keyup', debounce(200, e => {
    // TODO extract this and call it on page load too?
    if (e.target.value == '') {
      $('.node').removeClass('d-none')
      $('#search-input label').text('');
      return;
    }
    $('.node').addClass('d-none')
    const hits = idx.search(e.target.value)
    $('#search-input label').text(hits.length ? `${hits.length} matches` : 'no matches');
    for (hit of hits) {
      $(`#${hit.ref}`).removeClass('d-none');
    }
  }));

  return idx;
}

// FIXME rename?
function boot() {
    // all event handlers must be defered because we replace the entire
    // #codex-root element upon update from WebSocket
    $('body').on('click', '.node-head', event => {
        if (event.target.tagName == 'A') return;
        const $target = $(event.target)
        $target.closest('.node').toggleClass('collapsed');
    });

    /*********** Navigation **********/
    // FIXME revisit from here onwards
    $('#codex-root').on('click', '#expand-all', () => {
        $('.node').removeClass('collapsed');
    });

    $('#codex-root').on('click', '#collapse-all', () => {
        // only first depth, it's annoying to expand all depths later
        $('.node.node-depth-0').addClass('collapsed');
    });

    $('#codex-root').on('click', 'a.channel-link', (event) => {
        event.preventDefault();
        window.location.hash = event.currentTarget.hash;
    });

    $('#codex-root').on('click', '.fold-control', (event) => {
        const $control = $(event.currentTarget);
        const $foldable = $control.next()
        $control.toggleClass('collapsed');
        $foldable.toggleClass('collapsed');
    });

    $('#codex-root').on('click', 'nav', (event) => {
        const $target = $(event.target);
        if ($target.closest('#nav-controls').length) {
            return;
        }
        if ($target.closest('.channel-link').length) {
            return;
        }
        $('nav').toggleClass('collapsed');
    });

    /*********** Highlight **********/
    // nuance: moving cursor through nodes, in and out of nodes within a node
    $('body').on('mousemove mouseenter', '.node', event => {
        $('.node').removeClass('highlight');
        $(event.target).addClass('highlight');
        $(event.target).parents('.node').addClass('highlight')
    });
    $('body').on('mouseleave', '.node', event => {
        $(event.target).children('.node').removeClass('highlight');
        $(event.target).removeClass('highlight');
    });


    /*********** Full Screen ********/
    $('body').on('click', '.full-screen-button', event => {
        $('.node').removeClass('highlight');
        $('body').addClass('full-screen');
        $('#full-screen-modal').empty();
        $('#full-screen-modal').append($(event.target).closest('.node')[0].innerHTML);
        $('#full-screen-modal').removeClass('inactive');
    });
    $('html').on('click', event => {
        if (!$('body').hasClass('full-screen')) {
            return;
        }
        if ($(event.target).closest('.full-screen-button').length) {
            return;
        }
        if ($(event.target).closest('#full-screen-modal').length) {
            return;
        }
        $('body').removeClass('full-screen');
        $('#full-screen-modal').addClass('inactive');
    });
    $(document).keyup(event => {
        if (event.key === 'Escape') {
            $('body').removeClass('full-screen');
            $('#full-screen-modal').addClass('inactive');
        }
    });

    /*********** WebSocket Client ********/
    return; // HACK
    const ws = new WebSocket('ws://localhost:8080'); // FIXME variable port
    ws.onmessage = async (msg) => {
        const data = await msg.data;
        text = (typeof data === 'string') ? data : await data.text();
        if (text === '__UPDATING__') {
            console.log('Waiting for WebSocket to submit markup');
            $('#spinner').addClass('enabled');
            return;
        }

        console.log('Updating DOM from WebSocket');
        const parser = new DOMParser();
        const $newDoc = $(parser.parseFromString(text, 'text/html'));

        // remember which nodes were previously collapsed
        $('.node.collapsed').each((i, elem) => {
            if (elem.id) {
                $newDoc.find('#' + elem.id).addClass('collapsed');
            }
        });
        if ($('nav').hasClass('collapsed')) {
            $newDoc.find('nav').addClass('collapsed');
        } else {
            $newDoc.find('nav').removeClass('collapsed');
        }

        $('#codex-root').html($newDoc.find('#codex-root').html());
        // ad-hoc fix, see note in boot()
        $('#codex-root').append('<div id="full-screen-modal" class="inactive"></div>');
        // alert('Done setting #codex-root');
        $('head').html($newDoc.find('head')); // pandoc styles in <head> can change

        delete $newDoc, parser, text;

        $('#spinner').removeClass('enabled');
    }
}

$(document).ready(() => {
    boot();
    // FIXME this is a hack and probably breaks under websocket updates. But I
    // like moving all the html munging to FE!
    $('.node').each((idx, elem) => {
        $(elem).append(`<div class="full-screen-button"> ⤢ </div>`);
    });
    window.IDX = bootSearch();
    // WOOOO! it works
    // TODO pick up from here
    // 1. evaluate UI/UX
    // 2. clean up everything else, revisit "channels" (throw out?)
});
