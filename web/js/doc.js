// doc.js — peel mixin (classic script, no modules).
function peelDoc() {
  return {
    async loadFiles() {
      try {
        const r = await fetch('/api/files');
        if (!r.ok) return;
        const j = await r.json();
        this.files = j.files || [];
        this.root = j.root || 'docs';
      } catch (_) { /* keep previous list */ }
    },

    async openFile(path, opts = {}) {
      if (!path) return;
      const r = await fetch('/api/doc?path=' + encodeURIComponent(path));
      if (!r.ok) return;
      const d = await r.json();
      this.current = path;
      this.metaOpen = null;
      this.doc = d;
      this.slideIdx = 0;
      this.levelIdx = 0;
      if (d.type === 'deck' || d.type === 'c4') {
        this.leftOpen = false;
        this.rightOpen = false;
      }
      this.selQuote = '';
      this.composer.open = false;
      this.composerNote = '';
      this.clearC4Focus();
      clearTimeout(this._c4ClickT);
      const c = this.$refs.content;
      c.innerHTML = '';
      if (this.$refs.c4canvas) this.$refs.c4canvas.innerHTML = '';

      if (d.type === 'c4') {
        this.section = '';
        const start = Math.max(0, Math.min(d.levels.length - 1, (opts.level || 1) - 1));
        this.levelIdx = start;
        this.renderC4();
        this.$nextTick(() => this.c4ZoomFit());
      } else if (d.type === 'deck') {
        const start = Math.max(0, Math.min(d.slides.length - 1, (opts.slide || 1) - 1));
        d.slides.forEach((s, i) => {
          const div = document.createElement('div');
          div.className = 'slide' + (i === start ? ' active' : '');
          div.id = 'slide-' + i;
          div.innerHTML = s;
          const h = div.querySelector('h1,h2,h3');
          const strip = document.createElement('button');
          strip.type = 'button';
          strip.className = 'strip';
          strip.title = 'Go to slide ' + (i + 1);
          const n = document.createElement('span');
          n.className = 'strip-n';
          n.textContent = String(i + 1);
          const t = document.createElement('span');
          t.className = 'strip-t';
          t.textContent = h ? h.textContent : 'Slide ' + (i + 1);
          strip.append(n, t);
          strip.addEventListener('click', () => this.showSlide(i));
          div.prepend(strip);
          div.addEventListener('click', (ev) => {
            if (ev.target.closest('a,button')) return;
            if (window.getSelection().isCollapsed) this.showSlide(i);
          });
          c.appendChild(div);
        });
        wrapTables(c);
        mountDiagrams(c);
        this.showSlide(start, false);
      } else {
        c.innerHTML = d.html;
        this.slideTitle = '';
        wrapTables(c);
        mountDiagrams(c);
      }
      this.setupSpy();
      this.loadFeedback();
      this.syncURL();
    },

    setupSpy() {
      if (this._spy) { this._spy.disconnect(); this._spy = null; }
      if (this._slideSpy) { this._slideSpy.disconnect(); this._slideSpy = null; }
      if (this.doc?.type === 'c4') { this.section = ''; return; }
      if (this.doc?.type === 'deck') {
        const slides = [...this.$refs.content.querySelectorAll('.slide')];
        if (!slides.length) return;
        this._slideSpy = new IntersectionObserver(es => {
          if (this._slideLock && Date.now() < this._slideLock) return;
          let best = null, bestR = 0;
          es.forEach(en => { if (en.isIntersecting && en.intersectionRatio > bestR) { bestR = en.intersectionRatio; best = en.target; } });
          if (!best) return;
          const i = slides.indexOf(best);
          if (i >= 0 && i !== this.slideIdx) {
            this.slideIdx = i;
            const h = best.querySelector('h1,h2,h3');
            this.slideTitle = h ? h.textContent : 'Slide ' + (i + 1);
            this.syncURL();
          }
        }, { rootMargin: '-30% 0px -60% 0px', threshold: [0, 0.25, 0.5, 1] });
        slides.forEach(s => this._slideSpy.observe(s));
        return;
      }
      const heads = [...this.$refs.content.querySelectorAll('h1,h2,h3')];
      if (!heads.length) { this.section = ''; return; }
      this._spy = new IntersectionObserver(es => {
        es.forEach(en => { if (en.isIntersecting) this.section = en.target.textContent; });
      }, { rootMargin: '-20% 0px -70% 0px' });
      heads.forEach(h => this._spy.observe(h));
    },

    commentOnSection() {
      const heads = [...this.$refs.content.querySelectorAll('h1,h2,h3')];
      if (!heads.length) return;
      let i = heads.findIndex(h => h.textContent === this.section);
      if (i < 0) i = 0;
      const next = heads.slice(i + 1).find(h => h.tagName <= heads[i].tagName);
      const r = document.createRange();
      r.setStartBefore(heads[i]);
      if (next) r.setEndBefore(next);
      else if (this.$refs.content.lastChild) r.setEndAfter(this.$refs.content.lastChild);
      const s = window.getSelection();
      s.removeAllRanges();
      s.addRange(r);
      this.selQuote = this.selectedQuote();
      this.$refs.note.focus();
    },

    scrollToId(id) {
      const t = document.getElementById(id);
      if (t) t.scrollIntoView({ behavior: 'smooth' });
    },

    jumpToHeading(h) {
      this.metaOpen = null;
      if (!h) return;
      if (this.doc?.type === 'c4') {
        const i = (this.doc.levels || []).findIndex(l => l.name === h.text);
        this.showLevel(i >= 0 ? i : 0);
        return;
      }
      if (this.doc?.type === 'deck') {
        const slides = [...this.$refs.content.querySelectorAll('.slide')];
        for (let i = 0; i < slides.length; i++) {
          const m = slides[i].querySelector('h1,h2,h3');
          if (m && m.textContent.trim() === h.text.trim()) {
            this.showSlide(i);
            return;
          }
        }
        return;
      }
      this.scrollToId(h.id);
    },

    searchRoot() {
      return (this.doc?.type === 'c4' && this.$refs.c4canvas) ? this.$refs.c4canvas : this.$refs.content;
    },

    selectedQuote() {
      const sel = window.getSelection();
      if (!sel || sel.isCollapsed || sel.rangeCount === 0) return '';
      const r = sel.getRangeAt(0);
      const root = this.searchRoot();
      if (!root || !root.contains(r.commonAncestorContainer)) return '';
      return sel.toString().trim();
    },

    onSelect() {
      let q = this.selectedQuote();
      if (q && q.replace(/\s+/g, '').length < 3) q = '';
      const inNote = this.$refs.note && this.$refs.note.contains(document.activeElement);
      const inCmp = document.querySelector('#composer textarea') && document.querySelector('#composer textarea').contains(document.activeElement);
      if (!q) {
        if (this.composer.open && !inNote && !inCmp) this.closeComposer();
        return;
      }
      if (q && !inNote && !inCmp) {
        const c4n = this.captureC4FromSelection();
        if (c4n && (/[«»]|<<|\n/.test(q) || q.length > 80 || q.replace(/\s+/g, '').length < 2)) {
          q = c4n.label || c4n.id;
        }
        this.selQuote = q;
        this.composerQuote = q.length > 60 ? q.slice(0, 57) + '…' : q;
        const sel = window.getSelection();
        if (sel.rangeCount > 0) {
          const r = sel.getRangeAt(0).getBoundingClientRect();
          const pad = 8;
          let x = r.left;
          let y = r.bottom + pad;
          if (x + 320 > window.innerWidth) x = window.innerWidth - 330;
          if (y + 180 > window.innerHeight) y = Math.max(10, r.top - 190);
          this.composer = { open: true, x: Math.max(10, x), y };
        }
      } else if (!inNote && this.$refs.content.contains(document.activeElement)) {
        this.composer.open = false;
      }
    },
  };
}
