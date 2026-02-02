// Animate progress bar
document.addEventListener('DOMContentLoaded', () => {
  const progressFill = document.querySelector('.progress-fill');
  if (progressFill) {
    const progress = progressFill.getAttribute('data-progress');
    progressFill.style.setProperty('--progress-width', `${progress}%`);
  }

  // Add random movement to particles
  const particles = document.querySelectorAll('.particle');
  particles.forEach((particle, index) => {
    const randomX = (Math.random() - 0.5) * 200;
    const randomDelay = Math.random() * 5;
    
    particle.style.setProperty('--random-x', `${randomX}px`);
    particle.style.animationDelay = `${randomDelay}s`;
  });

  // Add subtle parallax effect on mouse move
  let mouseX = 0;
  let mouseY = 0;
  
  document.addEventListener('mousemove', (e) => {
    mouseX = (e.clientX / window.innerWidth - 0.5) * 20;
    mouseY = (e.clientY / window.innerHeight - 0.5) * 20;
    
    const logoIcon = document.querySelector('.logo-icon');
    if (logoIcon) {
      logoIcon.style.transform = `translate(${mouseX * 0.5}px, ${mouseY * 0.5}px)`;
    }
  });

  // Add glow effect on hover for cards
  const statCards = document.querySelectorAll('.stat-card');
  statCards.forEach(card => {
    card.addEventListener('mouseenter', () => {
      card.style.boxShadow = '0 0 30px rgba(139, 92, 246, 0.4)';
    });
    
    card.addEventListener('mouseleave', () => {
      card.style.boxShadow = 'none';
    });
  });
});

