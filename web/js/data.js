// data — default component state.
function peelData() {
  return {
    files: [], tab: 'all', current: null, doc: null, slideIdx: 0, levelIdx: 0, c4Zoom: 1, c4AutoFit: true,
    c4FocusId: '', c4FocusLabel: '', c4FocusFrom: '', c4FocusTo: '',
    root: 'docs', query: '', sortKey: 'path', sortDir: 1, userSorted: false, showAll: false,
    noteText: '', selQuote: '', noteAction: '', feedback: [], feedbackNotes: [], feedbackTotal: 0, fbAll: false, fbStatus: 'open', sse: 'connecting', section: '', slideTitle: '',
    agent: { state: 'idle', ref: '', note: '', since: '' },
    presets: [
      { icon: 'lightbulb', label: 'Explain this', action: 'explain' },
      { icon: 'wrench', label: 'Fix this', action: 'fix' },
      { icon: 'file-plus', label: 'Add example', action: 'example' },
      { icon: 'a-large-small', label: 'Simplify this', action: 'simplify' },
      { icon: 'circle-help', label: 'Question', action: 'question' },
    ],
    composer: { open: false, x: 0, y: 0 }, composerNote: '', composerAction: '', composerRect: null,
    leftOpen: false, rightOpen: false, metaOpen: null,
    goto: { open: false, query: '', idx: 0 },
    gotoHits: [],
    toasts: [],
    _toastSeq: 0,
  };
}
