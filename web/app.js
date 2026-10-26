// Alpine component for peel. Loaded as a classic script (end of body) so it
// registers on alpine:init before the deferred Alpine bundle initializes.
function peelApp() {
  return {
    files: [], tab: 'all', current: null, doc: null, slideIdx: 0, levelIdx: 0, c4Zoom: 1, c4AutoFit: true,
    root: 'docs', query: '', sortKey: 'path', sortDir: 1, userSorted: false, showAll: false,
    noteText: '', selQuote: '', feedback: [], feedbackTotal: 0, fbAll: false, sse: 'connecting', section: '', slideTitle: '',
    presets: [
      { icon: 'lightbulb', label: 'Explain this' },
      { icon: 'wrench', label: 'Fix this' },
      { icon: 'file-plus', label: 'Add example' },
      { icon: 'a-large-small', label: 'Simplify this' },
      { icon: 'circle-help', label: 'Question' },
    ],
    composer: { open: false, x: 0, y: 0 }, composerNote: '', composerRect: null,
    leftOpen: false, rightOpen: false, metaOpen: null,
    goto: { open: false, query: '', idx: 0 },
    gotoHits: [],

    typeIcon(t) {
      return t === 'c4' ? '[c4]' : (t === 'deck' ? '[dk]' : '[md]');
    },

    get crumbs() {
      if (!this.doc) return '—';
      if (this.doc.type === 'c4') return `${this.doc.path} · c4 · ${(this.doc.levels || []).length} levels`;
      if (this.doc.type === 'deck') return `${this.doc.path} · deck · ${(this.doc.slides || []).length} slides`;
      return this.doc.path;
    },
    get c4Crumb() {
      if (this.doc?.type !== 'c4') return '';
      const l = (this.doc.levels || [])[this.levelIdx];
      const name = l ? l.name : '';
      if (name && name === this.doc.title) return name;
      return `${this.doc.title} › ${name}`;
    },
    get headingLabel() {
      if (!this.doc) return 'headings ▾';
      if (this.doc.type === 'c4') {
        const n = (this.doc.levels || []).length;
        return `${n} ${n === 1 ? 'level' : 'levels'} ▾`;
      }
      if (this.doc.type === 'deck') {
        const n = (this.doc.slides || []).length;
        return `${n} ${n === 1 ? 'slide' : 'slides'} ▾`;
      }
      const n = (this.doc.outline || []).length;
      return `${n} ${n === 1 ? 'heading' : 'headings'} ▾`;
    },
    get backlinkLabel() {
      const n = this.doc ? (this.doc.backlinks || []).length : 0;
      return `${n} ${n === 1 ? 'backlink' : 'backlinks'} ▾`;
    },
    get visibleFiles() {
      const q = this.query.toLowerCase();
      const list = this.files.filter(f => {
        if (this.tab === 'docs' && f.type !== 'doc') return false;
        if (this.tab === 'decks' && f.type !== 'deck') return false;
        if (this.tab === 'c4' && f.type !== 'c4') return false;
        if (q && !(f.path.toLowerCase().includes(q) || (f.title || '').toLowerCase().includes(q))) return false;
        return true;
      });
      if (this.userSorted) {
        const val = f => this.sortKey === 'links' ? (f.links + f.backlinks)
          : this.sortKey === 'type' ? f.type : f.path;
        list.sort((a, b) => {
          const x = val(a), y = val(b);
          return (x < y ? -1 : x > y ? 1 : 0) * this.sortDir;
        });
      } else {
        list.sort((a, b) => String(b.modTime || '').localeCompare(String(a.modTime || '')));
      }
      return this.showAll ? list : list.slice(0, 5);
    },

    init() {
      const stored = localStorage.getItem('peel-theme');
      document.documentElement.dataset.theme = stored || 'dark';
      if (localStorage.getItem('peel-left') === '1') this.leftOpen = true;
      if (localStorage.getItem('peel-right') === '1') this.rightOpen = true;
      document.addEventListener('selectionchange', () => this.onSelect());
      document.addEventListener('keydown', e => {
        const mod = e.ctrlKey || e.metaKey;
        const key = e.key.toLowerCase();
        if (mod && (key === 'p' || key === 'k')) {
          e.preventDefault();
          this.openGoto();
          return;
        }
        if (this.goto.open) return;
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') {
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
          if (e.key === 'ArrowRight') step(cur + 1);
          if (e.key === 'ArrowLeft') step(cur - 1);
          if (e.key === 'Home') step(0);
          if (e.key === 'End') step(max);
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
        if (this.doc?.type === 'c4' && this.c4AutoFit) this.c4ZoomFit();
      });
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
    },

    toggleTheme() {
      const el = document.documentElement;
      el.dataset.theme = el.dataset.theme === 'dark' ? 'light' : 'dark';
      localStorage.setItem('peel-theme', el.dataset.theme);
    },
    toggleLeft() { this.leftOpen = !this.leftOpen; localStorage.setItem('peel-left', this.leftOpen ? '1' : '0'); },
    toggleRight() { this.rightOpen = !this.rightOpen; localStorage.setItem('peel-right', this.rightOpen ? '1' : '0'); },
    closeLeft() { this.leftOpen = false; localStorage.setItem('peel-left', '0'); },
    navFile(path) { this.openFile(path); this.closeLeft(); },

    // Sublime-style sequential fuzzy score; 0 = no match.
    fuzzyScore(text, query) {
      const t = String(text || '').toLowerCase();
      const q = String(query || '').toLowerCase();
      if (!q) return 1;
      let ti = 0, score = 0, run = 0, first = -1;
      for (let qi = 0; qi < q.length; qi++) {
        const ch = q[qi];
        let found = -1;
        for (let j = ti; j < t.length; j++) {
          if (t[j] === ch) { found = j; break; }
        }
        if (found < 0) return 0;
        if (first < 0) first = found;
        run = found === ti ? run + 1 : 1;
        score += 8 + run * 6;
        if (found === 0 || /[\/\-_.]/.test(t[found - 1])) score += 18;
        ti = found + 1;
      }
      score += Math.max(0, 40 - first);
      score += Math.max(0, 30 - (t.length - q.length));
      return score;
    },
    fileParts(path) {
      const slash = path.lastIndexOf('/');
      if (slash < 0) return { name: path, dir: '' };
      return { name: path.slice(slash + 1), dir: path.slice(0, slash) };
    },
    rankGotoHits() {
      const q = (this.goto.query || '').trim();
      let hits;
      if (!q) {
        hits = [...this.files]
          .sort((a, b) => String(b.modTime || '').localeCompare(String(a.modTime || '')))
          .slice(0, 24)
          .map(f => {
            const parts = this.fileParts(f.path);
            return { path: f.path, type: f.type, title: f.title, name: parts.name, dir: parts.dir, score: '·' };
          });
      } else {
        hits = [];
        for (const f of this.files) {
          const parts = this.fileParts(f.path);
          const base = this.fuzzyScore(parts.name, q);
          const full = this.fuzzyScore(f.path, q);
          const title = this.fuzzyScore(f.title || '', q);
          // Prefer basename matches.
          const score = Math.max(base * 2, full, title);
          if (score <= 0) continue;
          hits.push({ path: f.path, type: f.type, title: f.title, name: parts.name, dir: parts.dir, score });
        }
        hits.sort((a, b) => b.score - a.score || a.path.localeCompare(b.path));
        hits = hits.slice(0, 40);
      }
      this.gotoHits = hits;
      if (this.goto.idx >= hits.length) this.goto.idx = Math.max(0, hits.length - 1);
    },
    openGoto() {
      if (this.composer.open) this.closeComposer();
      this.goto.open = true;
      this.goto.query = '';
      this.goto.idx = 0;
      this.rankGotoHits();
      this.$nextTick(() => {
        const el = this.$refs.gotoInput;
        if (el) { el.focus(); el.select(); }
      });
    },
    closeGoto() {
      this.goto.open = false;
      this.goto.query = '';
      this.goto.idx = 0;
      this.gotoHits = [];
    },
    onGotoInput() {
      this.goto.idx = 0;
      this.rankGotoHits();
    },
    gotoMove(delta) {
      const n = this.gotoHits.length;
      if (!n) return;
      this.goto.idx = (this.goto.idx + delta + n) % n;
      this.$nextTick(() => {
        const on = document.querySelector('#gotoList li.on');
        if (on) on.scrollIntoView({ block: 'nearest' });
      });
    },
    gotoSelect() {
      const hit = this.gotoHits[this.goto.idx];
      if (hit) this.gotoOpen(hit.path);
    },
    gotoOpen(path) {
      this.closeGoto();
      this.openFile(path);
    },

    openFiles(tab) {
      this.tab = tab;
      this.leftOpen = true;
      localStorage.setItem('peel-left', '1');
      setTimeout(() => this.$refs.filter && this.$refs.filter.focus(), 50);
    },
    openAnnotate() {
      this.toggleRight();
      if (this.rightOpen) setTimeout(() => this.$refs.note && this.$refs.note.focus(), 50);
    },
    openFeedbackAll() {
      this.rightOpen = true;
      localStorage.setItem('peel-right', '1');
      this.loadAllFeedback();
    },
    toggleMeta(name) { this.metaOpen = this.metaOpen === name ? null : name; },
    syncURL() {
      if (!this.current) return;
      let u = '?file=' + encodeURIComponent(this.current);
      if (this.doc?.type === 'deck') u += '&slide=' + (this.slideIdx + 1);
      if (this.doc?.type === 'c4') u += '&level=' + (this.levelIdx + 1);
      history.replaceState(null, '', u);
    },
    setTab(t) { this.tab = t; },
    sort(k) { this.userSorted = true; if (this.sortKey === k) this.sortDir *= -1; else { this.sortKey = k; this.sortDir = 1; } },
    goHome() {
      const base = p => p.split('/').pop().toLowerCase();
      const rank = f => {
        const p = f.path.toLowerCase(), b = base(f.path);
        if (p === 'docs/index.md' || p === 'index.md') return 0;
        if (b === 'readme.md') return 1;
        return 2;
      };
      const cands = this.files.filter(f => rank(f) < 2).sort((a, b) => rank(a) - rank(b));
      if (cands.length) this.openFile(cands[0].path);
      else if (this.files.length) this.openFile(this.files[0].path);
    },

    async loadFiles() {
      const r = await fetch('/api/files');
      const j = await r.json();
      this.files = j.files || [];
      this.root = j.root || 'docs';
    },

    async openFile(path, opts = {}) {
      this.current = path;
      this.metaOpen = null;
      const r = await fetch('/api/doc?path=' + encodeURIComponent(path));
      if (!r.ok) return;
      const d = await r.json();
      this.doc = d;
      this.slideIdx = 0;
      this.levelIdx = 0;
      if (d.type === 'deck' || d.type === 'c4') {
        this.leftOpen = false;
        this.rightOpen = false;
        localStorage.setItem('peel-left', '0');
        localStorage.setItem('peel-right', '0');
      }
      this.selQuote = '';
      this.composer.open = false;
      this.composerNote = '';
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
          div.addEventListener('click', () => {
            if (window.getSelection().isCollapsed) this.showSlide(i);
          });
          c.appendChild(div);
        });
        mountDiagrams(c);
        this.showSlide(start, false);
      } else {
        c.innerHTML = d.html;
        this.slideTitle = '';
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

    showLevel(i) {
      if (this.doc?.type !== 'c4') return;
      const n = (this.doc.levels || []).length;
      this.levelIdx = Math.max(0, Math.min(n - 1, i));
      this.c4AutoFit = true;
      this.renderC4();
      this.$nextTick(() => this.c4ZoomFit());
      this.syncURL();
    },

    renderC4() {
      const box = this.$refs.c4canvas;
      if (!box || this.doc?.type !== 'c4') return;
      const l = (this.doc.levels || [])[this.levelIdx];
      box.innerHTML = '';
      this._c4NatW = 0;
      if (!l) return;
      const div = document.createElement('div');
      div.className = 'pele-diagram c4-diagram';
      box.appendChild(div);
      try { window.pele.mount(div, l.source.trim()); }
      catch (e) { div.textContent = 'diagram error: ' + e.message; }
      this.applyC4Zoom();
    },

    applyC4Zoom() {
      const box = this.$refs.c4canvas;
      const svg = box ? box.querySelector('svg') : null;
      if (!svg) return;
      if (!this._c4NatW) {
        this._c4NatW = parseFloat(svg.getAttribute('width')) || 0;
        if (!this._c4NatW && svg.viewBox && svg.viewBox.baseVal) this._c4NatW = svg.viewBox.baseVal.width;
        if (!this._c4NatW) this._c4NatW = svg.getBoundingClientRect().width || 800;
      }
      svg.style.width = (this._c4NatW * this.c4Zoom) + 'px';
      svg.style.height = 'auto';
      svg.style.maxWidth = 'none';
    },

    c4ZoomBy(d) {
      this.c4AutoFit = false;
      this.c4Zoom = Math.max(0.2, Math.min(3, this.c4Zoom * (d > 0 ? 1.2 : 1 / 1.2)));
      this.applyC4Zoom();
    },
    c4ZoomReset() { this.c4AutoFit = false; this.c4Zoom = 1; this.applyC4Zoom(); },
    c4ZoomFit() {
      const view = this.$refs.c4scroll;
      if (!view || !this._c4NatW) return;
      this.c4AutoFit = true;
      this.c4Zoom = Math.max(0.2, Math.min(3, (view.clientWidth - 48) / this._c4NatW));
      this.applyC4Zoom();
    },

    c4Drill(ev) {
      if (this.doc?.type !== 'c4') return;
      const g = ev.target && ev.target.closest ? ev.target.closest('.pele-node') : null;
      if (!g || !g.dataset.id) return;
      if (this.levelIdx + 1 < (this.doc.levels || []).length) this.showLevel(this.levelIdx + 1);
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

    showSlide(i, scroll = true) {
      if (this.doc?.type !== 'deck') return;
      const n = (this.doc.slides || []).length;
      this.slideIdx = Math.max(0, Math.min(n - 1, i));
      const slides = [...this.$refs.content.querySelectorAll('.slide')];
      slides.forEach((s, idx) => s.classList.toggle('active', idx === this.slideIdx));
      const cur = slides[this.slideIdx];
      if (cur) {
        mountDiagrams(cur);
        const h = cur.querySelector('h1,h2,h3');
        this.slideTitle = h ? h.textContent : 'Slide ' + (this.slideIdx + 1);
        if (scroll) {
          this._slideLock = Date.now() + 600;
          cur.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
        }
      }
      this.syncURL();
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
      const q = this.selectedQuote();
      const inNote = this.$refs.note && this.$refs.note.contains(document.activeElement);
      const inCmp = document.querySelector('#composer textarea') && document.querySelector('#composer textarea').contains(document.activeElement);
      if (q && !inNote && !inCmp) {
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
    closeComposer() { this.composer.open = false; this.composerNote = ''; },

    applyPreset(cur, preset) {
      const t = (cur || '').trim();
      if (!t) return preset + ': ';
      for (const p of this.presets) {
        const name = p.label;
        if (t === name) return preset + ': ';
        if (t.startsWith(name + ':')) return preset + t.slice(name.length) || preset + ': ';
        if (t.startsWith(name + ' ')) return preset + ': ' + t.slice(name.length).trim();
      }
      if (t.toLowerCase().startsWith(preset.toLowerCase())) return t;
      return preset + ': ' + t;
    },
    quickNote(p) {
      this.noteText = this.applyPreset(this.$refs.note ? this.$refs.note.value : this.noteText, p);
      this.$nextTick(() => this.$refs.note && this.$refs.note.focus());
    },
    quickComposer(p) {
      const box = document.querySelector('#composer textarea');
      this.composerNote = this.applyPreset(this.composerNote || (box ? box.value : ''), p);
      this.$nextTick(() => box && box.focus());
    },

    lineRangeFor(quote) {
      if (!this.doc || !quote) return [0, 0];
      const lines = (this.doc.body || '').split('\n');
      const first = quote.split('\n')[0].slice(0, 80);
      let start = 0;
      for (let i = 0; i < lines.length; i++) {
        if (first && lines[i].includes(first.slice(0, 24))) { start = i + 1; break; }
      }
      return [start, start + (quote.match(/\n/g) || []).length];
    },

    async saveComposer() {
      const box = document.querySelector('#composer textarea');
      const note = (this.composerNote || (box ? box.value : '')).trim();
      if (!this.selQuote || !note) return;
      const [sl, el] = this.lineRangeFor(this.selQuote);
      const body = { file: this.current, quote: this.selQuote, startLine: sl, endLine: el, note: this.composerNote.trim() };
      const r = await fetch('/api/feedback', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
      if (r.ok) { this.composerNote = ''; this.composer.open = false; this.selQuote = ''; window.getSelection().removeAllRanges(); this.loadFeedback(); this.refreshTotal(); }
    },

    timeAgo(ts) {
      const t = Date.parse(ts);
      if (isNaN(t)) return '';
      const s = Math.max(1, Math.floor((Date.now() - t) / 1000));
      if (s < 60) return s + 's ago';
      const m = Math.floor(s / 60);
      if (m < 60) return m + 'm ago';
      const h = Math.floor(m / 60);
      if (h < 24) return h + 'h ago';
      return Math.floor(h / 24) + 'd ago';
    },

    findBlockForQuote(quote) {
      const c = this.searchRoot();
      if (!c || !quote) return null;
      const anchor = quote.replace(/\s+/g, ' ').trim();
      if (anchor.length < 4) return null;
      const probes = [anchor.slice(0, 60), anchor.slice(0, 40), anchor.slice(0, 20), anchor.slice(0, 10)]
        .filter((p, idx, arr) => p.length >= 4 && arr.indexOf(p) === idx);
      for (const probe of probes) {
        const walker = document.createTreeWalker(c, NodeFilter.SHOW_TEXT);
        let node;
        while ((node = walker.nextNode())) {
          if (node.textContent.replace(/\s+/g, ' ').includes(probe)) {
            const host = node.parentElement && node.parentElement.closest ? node.parentElement : null;
            if (!host) return null;
            return host.closest('p,li,h1,h2,h3,h4,blockquote,td,.slide')
              || host.closest('.pele-diagram,pre');
          }
        }
      }
      return null;
    },

    jumpToFeedback(f) {
      if (!f || this.current !== f.file) { this.openFile(f.file); setTimeout(() => this.jumpToFeedback(f), 600); return; }
      if (this.doc?.type === 'deck') {
        const probe = this.findBlockForQuote(f.quote);
        const slideEl = probe ? probe.closest('.slide') : null;
        if (slideEl) {
          const i = [...this.$refs.content.querySelectorAll('.slide')].indexOf(slideEl);
          if (i >= 0 && i !== this.slideIdx) this.showSlide(i, false);
        }
      }
      if (this.doc?.type === 'c4') {
        const n = (this.doc.levels || []).length;
        for (let k = 0; k < n; k++) {
          this.showLevel(k);
          if (this.findBlockForQuote(f.quote)) break;
        }
      }
      const block = this.findBlockForQuote(f.quote);
      if (!block) return;
      block.scrollIntoView({ behavior: 'smooth', block: 'center' });
      block.classList.remove('fb-flash');
      void block.offsetWidth;
      block.classList.add('fb-flash');
      setTimeout(() => block.classList.remove('fb-flash'), 1400);
    },

    async deleteFeedback(f) {
      if (!f) return;
      await fetch('/api/feedback?id=' + encodeURIComponent(f.id), { method: 'DELETE' });
      if (this.fbAll) this.loadAllFeedback(); else this.loadFeedback();
      this.refreshTotal();
    },

    applyMarks() {
      const c = this.searchRoot();
      if (!c) return;
      c.querySelectorAll('.fb-mark').forEach(m => m.classList.remove('fb-mark', 'm0', 'm1', 'm2'));
      const list = this.fbAll ? (this.feedback || []).filter(f => f.file === this.current) : (this.feedback || []);
      list.forEach((f, i) => {
        const block = this.findBlockForQuote(f.quote);
        if (block && c.contains(block)) block.classList.add('fb-mark', 'm' + (i % 3));
      });
    },

    async saveNote() {
      const q = this.selQuote || this.selectedQuote();
      const note = (this.noteText || (this.$refs.note ? this.$refs.note.value : '')).trim();
      if (!q || !note) return;
      const [sl, el] = this.lineRangeFor(q);
      const body = { file: this.current, quote: q, startLine: sl, endLine: el, note };
      const r = await fetch('/api/feedback', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
      if (r.ok) { this.noteText = ''; this.selQuote = ''; window.getSelection().removeAllRanges(); this.loadFeedback(); this.refreshTotal(); }
    },

    async loadFeedback() {
      if (!this.current) return;
      this.fbAll = false;
      const r = await fetch('/api/feedback?file=' + encodeURIComponent(this.current) + '&t=' + Date.now());
      const j = await r.json();
      this.feedback = j.feedback || [];
      this.$nextTick(() => this.applyMarks());
    },

    async loadAllFeedback() {
      this.fbAll = true;
      const r = await fetch('/api/feedback?t=' + Date.now());
      const j = await r.json();
      this.feedback = j.feedback || [];
      this.feedbackTotal = this.feedback.length;
      this.$nextTick(() => this.applyMarks());
    },

    async refreshTotal() {
      try {
        const r = await fetch('/api/feedback?t=' + Date.now());
        const j = await r.json();
        this.feedbackTotal = (j.feedback || []).length;
      } catch (_) { /* ignore */ }
    },

    connectSSE() {
      const es = new EventSource('/events');
      es.onopen = () => { this.sse = 'live'; };
      es.onerror = () => { this.sse = 'retrying'; };
      es.onmessage = e => {
        try {
          const m = JSON.parse(e.data);
          if (m.type === 'reload') {
            this.loadFiles();
            if (m.path === this.current) this.openFile(this.current);
          }
        } catch (_) { /* ignore */ }
      };
    },
  };
}

function mountDiagrams(root) {
  if (!window.pele) return;
  root.querySelectorAll('pre.diagram').forEach(pre => {
    const code = pre.querySelector('code');
    const src = code ? code.textContent : pre.textContent;
    const div = document.createElement('div');
    div.className = 'pele-diagram';
    pre.replaceWith(div);
    try { window.pele.mount(div, src.trim()); }
    catch (e) { div.textContent = 'diagram error: ' + e.message; }
  });
}

document.addEventListener('alpine:init', () => { Alpine.data('peelApp', peelApp); });