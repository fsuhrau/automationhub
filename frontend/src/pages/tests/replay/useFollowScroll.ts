import {RefObject, useEffect} from 'react';
import {scrollTargetOf} from './eventView';

interface FollowState {
    follow: boolean,
    // changes with the line that is current
    activeKey: string | null,
    playing: boolean,
}

// useFollowScroll keeps the current row of a list in view
export const useFollowScroll = (list: RefObject<HTMLElement | null>, row: RefObject<HTMLElement | null>, s: FollowState): void => {
    const {follow, activeKey, playing} = s;
    useEffect(() => {
        const box = list.current;
        const active = row.current;
        if (!follow || !active || !box) return;
        const top = scrollTargetOf({
            rowTop: active.offsetTop - box.offsetTop, rowHeight: active.clientHeight, scrollTop: box.scrollTop, height: box.clientHeight,
        });
        if (top !== null) box.scrollTo({top, behavior: playing ? 'smooth' : 'auto'});
    }, [list, row, activeKey, follow, playing]);
};
