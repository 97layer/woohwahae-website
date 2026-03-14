const RUNTIME_TIME_FORMAT = new Intl.DateTimeFormat('en-GB', {
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit',
  hour12: false,
  timeZone: 'Asia/Seoul',
});

export function formatRuntimeTimestamp(value) {
  const timestamp = new Date(value);
  if (!Number.isFinite(timestamp.getTime())) {
    return '--:--:-- KST';
  }
  return `${RUNTIME_TIME_FORMAT.format(timestamp)} KST`;
}
