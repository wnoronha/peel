// deck.js — peel mixin (classic script, no modules).
function peelDeck() {
  return {
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
          this._slideLock = Date.now() + UI_TIMINGS.slideLockMs;
          cur.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
        }
      }
      this.syncURL();
    },

  };
}
