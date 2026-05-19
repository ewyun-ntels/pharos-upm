import React, {useState, useEffect} from 'react';
import {formatLocalTime} from '@pharos/shared/lib/unitUtils';

function Clock() {
  const [currentTime, setCurrentTime] = useState(new Date());

  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(new Date());
    }, 1000);

    return () => {
      clearInterval(timer);
    };
  }, []);

  return (
    <span className="text-xs text-muted-foreground">{formatLocalTime(currentTime)}</span>
  );
}

export default Clock;
