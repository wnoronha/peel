// goto.js — peel mixin (classic script, no modules).
function peelGoto() {
  return {
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
      this.$nextTick(() => { const el = this.$refs.filter; if (el) el.focus(); });
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

  };
}
