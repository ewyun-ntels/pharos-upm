const minute = 60;
const hour = minute * 60;
const day = hour * 24;

function timeDuration(time1: number | string | Date, time2: number | string | Date): string {
  const prevDate = Math.floor(new Date(time1).getTime() / 1000);
  const currentDate = Math.floor(new Date(time2).getTime() / 1000);

  const diff = currentDate - prevDate;

  if (diff < 0) {
    return '0s';
  }

  if (diff < minute) {
    return (diff % minute) + 's';
  }

  if (diff < minute * 3) {
    return Math.trunc(diff / minute) + 'm' + (diff % minute) + 's';
  }

  if (diff < hour) {
    return Math.trunc(diff / minute) + 'm';
  }

  if (diff < 3 * hour) {
    return Math.trunc(diff / hour) + 'h' + Math.trunc((diff % hour) / minute) + 'm';
  }

  if (diff < day) {
    return Math.trunc(diff / hour) + 'h';
  }

  if (diff < day * 7) {
    return Math.trunc(diff / day) + 'd' + Math.trunc((diff % day) / hour) + 'h';
  }

  return Math.trunc(diff / day) + 'd';
}

export default timeDuration;
