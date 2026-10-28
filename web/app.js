// Alpine component for peel. Composition root — classic scripts in web/js/* load before this file.
function peelApp() {
  const o = {};
  for (const part of [peelData(), peelUtils(), peelState(), peelGoto(), peelDoc(), peelDeck(), peelC4(), peelFeedback(), peelSSE()]) {
    for (const [k, d] of Object.entries(Object.getOwnPropertyDescriptors(part))) {
      Object.defineProperty(o, k, d);
    }
  }
  Object.assign(o, {
    init() {
      const stored = localStorage.getItem('peel-theme');
      document.documentElement.dataset.theme = stored || 'dark';
      if (localStorage.getItem('peel-left') === '1') this.leftOpen = true;
      if (localStorage.getItem('peel-right') === '1') this.rightOpen = true;
      document.addEventListener('selectionchange', () => {
        if (this._selQueued) return;
        this._selQueued = true;
        requestAnimationFrame(() => {
          this._selQueued = false;
          try { this.onSelect(); } catch (_) { /* ignore */ }
        });
      });
      document.addEventListener('keydown', e => {
        const mod = e.ctrlKey || e.metaKey;
        const key = e.key.toLowerCase();
        if (mod && (key === 'p' || key === 'k')) {
          e.preventDefault();
          this.openGoto();
          return;
        }
        if (this.goto.open) return;
        const tag = (e.target.tagName || '').toUpperCase();
        if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || tag === 'BUTTON' || e.target.isContentEditable) {
          if (e.key === 'Escape') e.target.blur();
          return;
        }
        if (e.key === '[') this.toggleLeft();
        if (e.key === ']') this.toggleRight();
        if (this.doc && (this.doc.type === 'deck' || this.doc.type === 'c4')) {
          const isC4 = this.doc.type === 'c4';
          const cur = isC4 ? this.levelIdx : this.slideIdx;
          const max = (isC4 ? (this.doc.levels || []).length : (this.doc.slides || []).length) - 1;
          const step = isC4 ? i => this.showLevel(i) : i => this.showSlide(i);
          if (e.key === 'ArrowRight') { e.preventDefault(); step(cur + 1); }
          if (e.key === 'ArrowLeft') { e.preventDefault(); step(cur - 1); }
          if (e.key === 'Home') { e.preventDefault(); step(0); }
          if (e.key === 'End') { e.preventDefault(); step(max); }
        }
      });
      this.$refs.content.addEventListener('click', e => {
        const a = e.target.closest('a[data-wikilink], a[data-doclink]');
        if (a) {
          e.preventDefault();
          const target = a.dataset.wikilink || a.dataset.doclink;
          const href = a.getAttribute('href') || '';
          const hashIdx = href.indexOf('#');
          const hash = hashIdx >= 0 ? href.slice(hashIdx + 1) : '';
          this.openFile(target).then(() => {
            if (hash) {
              this.$nextTick(() => this.scrollToId(hash));
            }
          });
        }
      });
      window.addEventListener('resize', () => {
        if (this._resizeT) return;
        this._resizeT = setTimeout(() => {
          this._resizeT = null;
          if (this.doc?.type === 'c4' && this.c4AutoFit) this.c4ZoomFit();
        }, UI_TIMINGS.resizeMs);
      });
      window.addEventListener('scroll', () => {
        if (this.composer.open) this.closeComposer();
      }, { passive: true, capture: true });
      this.loadFiles().then(() => {
        const params = new URLSearchParams(location.search);
        const f = params.get('file') || params.get('view');
        const sn = parseInt(params.get('slide'), 10);
        const ln = parseInt(params.get('level'), 10);
        const opts = { slide: Number.isFinite(sn) ? sn : 1, level: Number.isFinite(ln) ? ln : 1 };
        if (f && this.files.some(x => x.path === f)) this.openFile(f, opts);
        else if (this.files.length) this.openFile(this.files[0].path);
      });
      this.connectSSE();
      this.refreshTotal();
      this.loadAgent();
    },

    }
  );
  return o;
}

document.addEventListener('alpine:init', () => { Alpine.data('peelApp', peelApp); });
