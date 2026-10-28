// state.js — peel mixin (classic script, no modules).
function peelState() {
  return {
    get agentStatusLabel() {
      if (this.agent.state !== 'working') return '';
      const n = (this.agent.note || '').trim();
      return n || 'Agent working…';
    },
    get agentWorkingTitle() {
      return this.agent.state === 'working' ? this.agentStatusLabel : 'peel home';
    },

    get crumbs() {
      if (!this.doc) return '—';
      if (this.doc.type === 'c4') return `${this.doc.path} · c4 · ${(this.doc.levels || []).length} levels`;
      if (this.doc.type === 'deck') return `${this.doc.path} · deck · ${(this.doc.slides || []).length} slides`;
      return this.doc.path;
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
    get filteredFiles() {
      const q = this.query.toLowerCase();
      const list = this.files.filter(f => {
        if (this.tab === 'docs' && f.type !== 'doc') return false;
        if (this.tab === 'decks' && f.type !== 'deck') return false;
        if (this.tab === 'c4' && f.type !== 'c4') return false;
        if (q && !(f.path.toLowerCase().includes(q) || (f.title || '').toLowerCase().includes(q))) return false;
        return true;
      });
      if (this.userSorted) {
        const val = f => this.sortKey === 'links' ? ((f.links || 0) + (f.backlinks || 0))
          : this.sortKey === 'type' ? f.type : f.path;
        list.sort((a, b) => {
          const x = val(a), y = val(b);
          return (x < y ? -1 : x > y ? 1 : 0) * this.sortDir;
        });
      } else {
        list.sort((a, b) => String(b.modTime || '').localeCompare(String(a.modTime || '')));
      }
      return list;
    },
    get visibleFiles() {
      return this.showAll ? this.filteredFiles : this.filteredFiles.slice(0, 5);
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

    syncURL() {
      if (!this.current) return;
      let u = '?file=' + encodeURIComponent(this.current);
      if (this.doc?.type === 'deck') u += '&slide=' + (this.slideIdx + 1);
      if (this.doc?.type === 'c4') u += '&level=' + (this.levelIdx + 1);
      history.replaceState(null, '', u);
    },
    toggleMeta(name) { this.metaOpen = this.metaOpen === name ? null : name; },
    openFiles(tab) {
      this.tab = tab;
      this.leftOpen = true;
      localStorage.setItem('peel-left', '1');
      setTimeout(() => this.$refs.filter && this.$refs.filter.focus(), 50);
    },
    openFeedbackAll() {
      this.rightOpen = true;
      localStorage.setItem('peel-right', '1');
      this.loadAllFeedback();
    },
  };
}
