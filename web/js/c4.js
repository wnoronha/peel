// c4.js — peel mixin (classic script, no modules).
function peelC4() {
  return {
    get c4Crumb() {
      if (this.doc?.type !== 'c4') return '';
      const l = (this.doc.levels || [])[this.levelIdx];
      const name = l ? l.name : '';
      if (name && name === this.doc.title) return name;
      return `${this.doc.title} › ${name}`;
    },
    get c4LevelName() {
      if (this.doc?.type !== 'c4') return '—';
      const l = (this.doc.levels || [])[this.levelIdx];
      return l ? l.name : '—';
    },
    get c4LevelKind() {
      if (this.doc?.type !== 'c4') return '';
      return this.c4Kind((this.doc.levels || [])[this.levelIdx]);
    },
    get c4LevelPos() {
      if (this.doc?.type !== 'c4') return '';
      const n = (this.doc.levels || []).length;
      if (!n) return '';
      return `${this.levelIdx + 1}/${n}`;
    },
    c4Kind(l) {
      if (!l) return '';
      if (l.kind) return l.kind;
      const m = String(l.source || '').match(/^\s*(C4Context|C4Container|C4Component|C4Code|C4Dynamic|C4Deployment)\b/m);
      if (!m) return '';
      return ({ C4Context: 'Context', C4Container: 'Container', C4Component: 'Component', C4Code: 'Code', C4Dynamic: 'Dynamic', C4Deployment: 'Deployment' })[m[1]] || '';
    },
    c4KindIcon(l) {
      const k = l ? this.c4Kind(l) : this.c4LevelKind;
      return ({ Context: 'layers', Container: 'box', Component: 'component', Code: 'code-xml', Deployment: 'server' })[k] || '';
    },
    c4LevelLabel(l) {
      if (!l) return '—';
      const k = this.c4Kind(l);
      return k ? `${k} · ${l.name}` : l.name;
    },
    showLevel(i) {
      if (this.doc?.type !== 'c4') return;
      const n = (this.doc.levels || []).length;
      this.levelIdx = Math.max(0, Math.min(n - 1, i));
      this.clearC4Focus();
      clearTimeout(this._c4ClickT);
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

    clearC4Focus() {
      this.c4FocusId = '';
      this.c4FocusLabel = '';
      this.c4FocusFrom = '';
      this.c4FocusTo = '';
    },

    c4LevelSource() {
      return ((this.doc && this.doc.levels || [])[this.levelIdx] || {}).source || (this.doc && this.doc.body) || '';
    },

    c4LabelForId(id) {
      if (!id || !this.doc) return '';
      const src = this.c4LevelSource();
      const esc = String(id).replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      const re = new RegExp(
        '\\b(?:Person|System_Ext|System|ContainerDb|Container_Ext|Container|Component_Ext|Component)\\(\\s*' +
          esc + '\\s*,\\s*"([^"]+)"'
      );
      const m = src.match(re);
      return m ? m[1] : '';
    },

    c4RelInfoForEdgeId(edgeId) {
      if (!edgeId || !this.doc) return null;
      const src = this.c4LevelSource();
      const re = /\bRel(?:_[UDLR])?\(\s*([A-Za-z0-9_-]+)\s*,\s*([A-Za-z0-9_-]+)\s*,\s*"([^"]*)"/g;
      let m;
      while ((m = re.exec(src))) {
        const from = m[1], to = m[2], label = m[3];
        if (from + '-' + to === edgeId) return { id: edgeId, from, to, label };
      }
      return null;
    },

    setC4FocusFromNode(g) {
      if (!g || !g.dataset || !g.dataset.id) return null;
      const id = g.dataset.id;
      const label = this.c4LabelForId(id) || '';
      this.c4FocusId = id;
      this.c4FocusLabel = label;
      this.c4FocusFrom = '';
      this.c4FocusTo = '';
      return { id, label, from: '', to: '' };
    },

    setC4FocusFromEdge(g) {
      if (!g || !g.dataset || !g.dataset.id) return null;
      const id = g.dataset.id;
      const info = this.c4RelInfoForEdgeId(id) || { id, from: '', to: '', label: '' };
      let label = (info.label || '').trim();
      if (!label && g.classList.contains('pele-edge-label')) {
        label = (g.textContent || '').replace(/\s+/g, ' ').trim();
      }
      this.c4FocusId = id;
      this.c4FocusLabel = label;
      this.c4FocusFrom = info.from || '';
      this.c4FocusTo = info.to || '';
      return { id, label, from: this.c4FocusFrom, to: this.c4FocusTo };
    },

    setC4FocusFromEl(el) {
      if (!el || !el.closest) return null;
      const node = el.closest('.pele-node');
      if (node) return this.setC4FocusFromNode(node);
      const edge = el.closest('.pele-edge, .pele-edge-label');
      if (edge) return this.setC4FocusFromEdge(edge);
      return null;
    },

    captureC4FromSelection() {
      if (this.doc?.type !== 'c4') return null;
      const sel = window.getSelection();
      if (!sel || !sel.rangeCount) return null;
      let el = sel.anchorNode;
      el = el && el.nodeType === 3 ? el.parentElement : el;
      return this.setC4FocusFromEl(el);
    },

    openC4Composer(ev, hit) {
      const q = hit.label || hit.id;
      this.selQuote = q;
      this.composerQuote = q.length > 60 ? q.slice(0, 57) + '…' : q;
      const pad = 8;
      let x = ev.clientX + pad, y = ev.clientY + pad;
      if (x + 320 > window.innerWidth) x = window.innerWidth - 330;
      if (y + 180 > window.innerHeight) y = Math.max(10, ev.clientY - 190);
      this.composer = { open: true, x: Math.max(10, x), y };
      this.$nextTick(() => {
        const box = document.querySelector('#composer textarea');
        if (box) box.focus();
      });
    },

    c4Comment(ev) {
      if (this.doc?.type !== 'c4') return;
      const hit = this.setC4FocusFromEl(ev.target);
      if (!hit) return;
      this.openC4Composer(ev, hit);
    },

    c4Drill(ev) {
      if (this.doc?.type !== 'c4') return;
      const g = ev.target && ev.target.closest ? ev.target.closest('.pele-node') : null;
      if (!g || !g.dataset.id) return;
      if (this.levelIdx + 1 < (this.doc.levels || []).length) this.showLevel(this.levelIdx + 1);
    },

    // Single-click comments; double-click drills (timer avoids firing both).
    c4Click(ev) {
      if (this.doc?.type !== 'c4') return;
      if (!this.setC4FocusFromEl(ev.target)) return;
      clearTimeout(this._c4ClickT);
      const x = ev.clientX, y = ev.clientY, t = ev.target;
      this._c4ClickT = setTimeout(() => {
        this.c4Comment({ clientX: x, clientY: y, target: t });
      }, 280);
    },

    c4DblClick(ev) {
      if (this.doc?.type !== 'c4') return;
      clearTimeout(this._c4ClickT);
      this.composer.open = false;
      this.c4Drill(ev);
    },

  };
}
