// feedback.js — peel mixin (classic script, no modules).
function peelFeedback() {
  return {
    closeComposer() {
      this.composer.open = false;
      this.composerNote = '';
      this.composerAction = '';
      this.clearC4Focus();
    },

    applyPreset(cur, preset) {
      const t = (cur || '').trim();
      if (!t) return preset + ': ';
      for (const p of this.presets) {
        const name = p.label;
        if (t === name) return preset + ': ';
        if (t.startsWith(name + ':')) return preset + (t.slice(name.length) || ': ');
        if (t.startsWith(name + ' ')) return preset + ': ' + t.slice(name.length).trim();
      }
      if (t.toLowerCase().startsWith(preset.toLowerCase())) return t;
      return preset + ': ' + t;
    },
    quickNote(p) {
      const preset = typeof p === 'string' ? this.presets.find(x => x.label === p) : p;
      const label = preset ? preset.label : p;
      if (preset && preset.action) this.noteAction = preset.action;
      this.noteText = this.applyPreset(this.$refs.note ? this.$refs.note.value : this.noteText, label);
      this.$nextTick(() => this.$refs.note && this.$refs.note.focus());
    },
    quickComposer(p) {
      const preset = typeof p === 'string' ? this.presets.find(x => x.label === p) : p;
      const label = preset ? preset.label : p;
      if (preset && preset.action) this.composerAction = preset.action;
      const box = document.querySelector('#composer textarea');
      this.composerNote = this.applyPreset(this.composerNote || (box ? box.value : ''), label);
      this.$nextTick(() => box && box.focus());
    },

    lineRangeFor(quote, c4Id, c4From, c4To) {
      if (!this.doc) return [0, 0];
      const lines = (this.doc.body || '').split('\n');
      if (c4From && c4To) {
        const esc = s => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
        const re = new RegExp('\\bRel[^\\n]*\\b' + esc(c4From) + '\\b[^\\n]*\\b' + esc(c4To) + '\\b|\\bRel[^\\n]*\\b' + esc(c4To) + '\\b[^\\n]*\\b' + esc(c4From) + '\\b');
        for (let i = 0; i < lines.length; i++) {
          if (re.test(lines[i])) return [i + 1, i + 1];
        }
      }
      if (c4Id) {
        const a = '(' + c4Id + ',';
        const b = '(' + c4Id + ')';
        for (let i = 0; i < lines.length; i++) {
          if (lines[i].includes(a) || lines[i].includes(b)) return [i + 1, i + 1];
        }
      }
      if (!quote) return [0, 0];
      const first = quote.split('\n')[0].slice(0, 80);
      let start = 0;
      for (let i = 0; i < lines.length; i++) {
        if (first && lines[i].includes(first.slice(0, 24))) { start = i + 1; break; }
      }
      return [start, start + (quote.match(/\n/g) || []).length];
    },

    // Stable view anchors alongside quote/lines (deck slide, C4 level + node/edge, md heading id).
    feedbackAnchor() {
      const a = {};
      if (!this.doc) return a;
      if (this.doc.type === 'deck') {
        a.slide = this.slideIdx + 1;
        return a;
      }
      if (this.doc.type === 'c4') {
        a.level = this.levelIdx + 1;
        if (this.c4LevelKind) a.levelKind = this.c4LevelKind;
        if (!this.c4FocusId) this.captureC4FromSelection();
        if (this.c4FocusId) {
          a.c4Id = this.c4FocusId;
          if (this.c4FocusLabel) a.c4Label = this.c4FocusLabel;
          if (this.c4FocusFrom) a.c4From = this.c4FocusFrom;
          if (this.c4FocusTo) a.c4To = this.c4FocusTo;
        }
        return a;
      }
      const outline = this.doc.outline || [];
      let id = '';
      const sel = window.getSelection();
      if (sel && sel.rangeCount && this.$refs.content) {
        let node = sel.anchorNode;
        let el = node && node.nodeType === 3 ? node.parentElement : node;
        while (el && el !== this.$refs.content) {
          if (el.id && /^H[1-6]$/.test(el.tagName)) { id = el.id; break; }
          el = el.previousElementSibling || (el.parentElement !== this.$refs.content ? el.parentElement : null);
        }
        if (!id) {
          let cur = sel.anchorNode && (sel.anchorNode.nodeType === 3 ? sel.anchorNode.parentElement : sel.anchorNode);
          while (cur && cur !== this.$refs.content) {
            const prev = cur.previousElementSibling;
            if (prev && /^H[1-6]$/.test(prev.tagName) && prev.id) { id = prev.id; break; }
            cur = cur.parentElement;
          }
        }
      }
      if (!id && this.section) {
        const hit = outline.find(o => o.text === this.section);
        if (hit && hit.id) id = hit.id;
      }
      if (id) a.headingId = id;
      return a;
    },

    makeNotePayload(rawQuote, note) {
      const anchor = this.feedbackAnchor();
      let quote = rawQuote;
      if (anchor.c4Id && (/[«»]|<<|\n/.test(quote) || quote.length > 80)) {
        quote = anchor.c4Label || anchor.c4Id;
      }
      const [sl, el] = this.lineRangeFor(quote, anchor.c4Id, anchor.c4From, anchor.c4To);
      return { quote, body: { type: 'note', file: this.current, quote, startLine: sl, endLine: el, note, ...anchor } };
    },

    async saveComposer() {
      if (this._saving) return;
      const box = document.querySelector('#composer textarea');
      const note = (this.composerNote || (box ? box.value : '')).trim();
      if (!this.selQuote || !note) return;
      this._saving = true;
      try {
      const made = this.makeNotePayload(this.selQuote, this.composerNote.trim());
      const body = made.body;
      if (this.composerAction) body.action = this.composerAction;
      const r = await fetch('/api/feedback', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
      if (r.ok) {
        try {
          const j = await r.json();
          if (j && j.id) {
            this._ownNoteIds = this._ownNoteIds || {};
            this._ownNoteIds[j.id] = Date.now();
          }
        } catch (_) { /* ignore */ }
        this.composerNote = '';
        this.composerAction = '';
        this.composer.open = false;
        this.selQuote = '';
        this.clearC4Focus();
        window.getSelection().removeAllRanges();
        this.loadFeedback();
        this.refreshTotal();
      } else {
        this.toast('Save failed', { kind: 'warn' });
      }
      } finally {
        this._saving = false;
      }
    },

    // Classify raw log events into notes with status: open | acked | dismissed.
    classifyNotes(events) {
      const closes = Object.create(null);
      for (const ev of events || []) {
        const t = ev.type || 'note';
        if ((t === 'ack' || t === 'tombstone') && ev.ref) {
          closes[ev.ref] = {
            status: t === 'ack' ? 'acked' : 'dismissed',
            ts: ev.ts,
            note: ev.note || '',
          };
        }
      }
      const out = [];
      for (const ev of events || []) {
        const t = ev.type || 'note';
        if (t !== 'note') continue;
        const c = closes[ev.id];
        out.push({
          ...ev,
          type: 'note',
          status: c ? c.status : 'open',
          closedAt: c ? c.ts : '',
          closeNote: c ? c.note : '',
        });
      }
      return out;
    },

    get visibleFeedback() {
      const st = this.fbStatus || 'open';
      let list = this.feedbackNotes || [];
      if (st === 'open') list = list.filter(f => f.status === 'open');
      else if (st === 'acked') list = list.filter(f => f.status === 'acked');
      else if (st === 'dismissed') list = list.filter(f => f.status === 'dismissed');
      return list;
    },

    fbStatusCount(status) {
      const list = this.feedbackNotes || [];
      if (status === 'all') return list.length;
      return list.filter(f => f.status === status).length;
    },

    setFbStatus(status) {
      this.fbStatus = status;
      this.$nextTick(() => this.applyMarks());
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

    async jumpToFeedback(f) {
      if (!f) return;
      if (this.current !== f.file) {
        await this.openFile(f.file);
        if (this.current !== f.file) return;
      }
      if (this.doc?.type === 'deck' && f.slide > 0) {
        this.showSlide(f.slide - 1, false);
      } else if (this.doc?.type === 'deck') {
        const probe = this.findBlockForQuote(f.quote);
        const slideEl = probe ? probe.closest('.slide') : null;
        if (slideEl) {
          const i = [...this.$refs.content.querySelectorAll('.slide')].indexOf(slideEl);
          if (i >= 0 && i !== this.slideIdx) this.showSlide(i, false);
        }
      }
      if (this.doc?.type === 'c4' && f.level > 0) {
        this.showLevel(f.level - 1);
      } else if (this.doc?.type === 'c4') {
        const levels = this.doc.levels || [];
        const q = (f.quote || '').slice(0, 60);
        let best = 0;
        if (q.length >= 4) {
          for (let k = 0; k < levels.length; k++) {
            if ((levels[k].Source || levels[k].source || '').includes(q)) { best = k; break; }
          }
        }
        this.showLevel(best);
      }
      if (f.headingId) {
        const el = this.$refs.content && this.$refs.content.querySelector('#' + CSS.escape(f.headingId));
        if (el) {
          el.scrollIntoView({ behavior: 'smooth', block: 'center' });
          el.classList.remove('fb-flash');
          void el.offsetWidth;
          el.classList.add('fb-flash');
          setTimeout(() => el.classList.remove('fb-flash'), UI_TIMINGS.flashMs);
          return;
        }
      }
      const block = this.findBlockForQuote(f.quote);
      if (!block) return;
      block.scrollIntoView({ behavior: 'smooth', block: 'center' });
      block.classList.remove('fb-flash');
      void block.offsetWidth;
      block.classList.add('fb-flash');
      setTimeout(() => block.classList.remove('fb-flash'), UI_TIMINGS.flashMs);
    },

    async deleteFeedback(f) {
      if (!f || f.status !== 'open') return;
      await fetch('/api/feedback?id=' + encodeURIComponent(f.id), { method: 'DELETE' });
      await this.reloadFeedback();
    },

    applyMarks() {
      const c = this.searchRoot();
      if (!c) return;
      c.querySelectorAll('.fb-mark').forEach(m => m.classList.remove('fb-mark', 'm0', 'm1', 'm2'));
      // Index text blocks once per pass instead of walking per note.
      const nodes = [];
      try {
        const walker = document.createTreeWalker(c, NodeFilter.SHOW_TEXT);
        let node;
        while ((node = walker.nextNode())) {
          const txt = (node.textContent || '').replace(/\s+/g, ' ');
          if (txt.length < 4) continue;
          const host = node.parentElement && node.parentElement.closest ? node.parentElement : null;
          const block = host ? (host.closest('p,li,h1,h2,h3,h4,blockquote,td,.slide') || host.closest('.pele-diagram,pre')) : null;
          if (block) nodes.push({ text: txt, block });
        }
      } catch (_) { /* ignore */ }
      const findInIndex = (quote) => {
        const anchor = (quote || '').replace(/\s+/g, ' ').trim();
        if (anchor.length < 4 || !nodes.length) return null;
        const probes = [anchor.slice(0, 60), anchor.slice(0, 40), anchor.slice(0, 20), anchor.slice(0, 10)]
          .filter((p, idx, arr) => p.length >= 4 && arr.indexOf(p) === idx);
        for (const probe of probes) {
          for (const n of nodes) {
            if (n.text.includes(probe)) return n.block;
          }
        }
        return null;
      };
      // Only open notes mark the doc — acked/dismissed stay in the list but don't paint.
      const list = (this.feedbackNotes || []).filter(f => f.status === 'open' && f.file === this.current);
      list.forEach((f, i) => {
        const block = findInIndex(f.quote);
        if (block && c.contains(block)) block.classList.add('fb-mark', 'm' + (i % 3));
      });
    },

    async saveNote() {
      if (this._saving) return;
      let q = this.selQuote || this.selectedQuote();
      const note = (this.noteText || (this.$refs.note ? this.$refs.note.value : '')).trim();
      if (!q || !note) return;
      this._saving = true;
      try {
      const made = this.makeNotePayload(q, note);
      q = made.quote;
      const body = made.body;
      if (this.noteAction) body.action = this.noteAction;
      const r = await fetch('/api/feedback', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
      if (r.ok) {
        try {
          const j = await r.json();
          if (j && j.id) {
            this._ownNoteIds = this._ownNoteIds || {};
            this._ownNoteIds[j.id] = Date.now();
          }
        } catch (_) { /* ignore */ }
        this.noteText = '';
        this.noteAction = '';
        this.selQuote = '';
        this.clearC4Focus();
        window.getSelection().removeAllRanges();
        this.loadFeedback();
        this.refreshTotal();
      } else {
        this.toast('Save failed', { kind: 'warn' });
      }
      } finally {
        this._saving = false;
      }
    },

    async reloadFeedback() {
      const r = await fetch('/api/feedback?all=1&t=' + Date.now());
      const j = await r.json();
      let notes = this.classifyNotes(j.feedback || []);
      if (!this.fbAll && this.current) notes = notes.filter(f => f.file === this.current);
      this.feedbackNotes = notes;
      this.feedback = notes.filter(f => f.status === 'open');
      this.feedbackTotal = (this.classifyNotes(j.feedback || [])).filter(f => f.status === 'open').length;
      this.$nextTick(() => this.applyMarks());
    },

    async loadFeedback() {
      if (!this.current) return;
      this.fbAll = false;
      await this.reloadFeedback();
    },

    async loadAllFeedback() {
      this.fbAll = true;
      await this.reloadFeedback();
    },

    async refreshTotal() {
      try {
        const r = await fetch('/api/feedback?all=1&t=' + Date.now());
        const j = await r.json();
        this.feedbackTotal = this.classifyNotes(j.feedback || []).filter(f => f.status === 'open').length;
      } catch (_) { /* ignore */ }
    },

  };
}
