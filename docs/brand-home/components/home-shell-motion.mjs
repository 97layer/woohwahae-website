export const PULL_REFRESH_BASE_FRICTION = 0.46;
export const PULL_REFRESH_MIN_FRICTION = 0.16;
export const PULL_REFRESH_FRICTION_DROP_PER_PIXEL = 0.001;
export const PULL_REFRESH_MAX_DISTANCE = 92;
export const PULL_REFRESH_READY_DISTANCE = 62;
export const PULL_REFRESH_SETTLED_DISTANCE = 54;
export const PULL_REFRESH_RELOAD_DELAY_MS = 180;

export const ARCHITECTURAL_EASING = 'cubic-bezier(0.16, 1, 0.3, 1)';

/**
 * DNA: Whimsy Tension Curve
 * A soft, spring-like response for UI elements.
 */
export function computeWhimsyTension(current, target, friction = 0.15) {
  return current + (target - current) * friction;
}

/**
 * DNA: Magnetic Pull
 * Calculates a subtle pull towards a point with a specific threshold.
 */
export function computeMagneticPull(x, y, rect, strength = 0.3, range = 100) {
  const centerX = rect.left + rect.width / 2;
  const centerY = rect.top + rect.height / 2;
  const dx = x - centerX;
  const dy = y - centerY;
  const distance = Math.sqrt(dx * dx + dy * dy);

  if (distance < range) {
    const power = (1 - distance / range) * strength;
    return { x: dx * power, y: dy * power };
  }
  return { x: 0, y: 0 };
}

export function computePullRefreshDistance(deltaY) {
  if (deltaY <= 0) return 0;

  const friction = Math.max(
    PULL_REFRESH_MIN_FRICTION,
    PULL_REFRESH_BASE_FRICTION - (deltaY * PULL_REFRESH_FRICTION_DROP_PER_PIXEL),
  );

  return Math.min(PULL_REFRESH_MAX_DISTANCE, deltaY * friction);
}

export function isPullRefreshReady(distance) {
  return distance >= PULL_REFRESH_READY_DISTANCE;
}
