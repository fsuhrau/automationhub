import React from 'react';
import {Section} from '../../../components/common';
import {AndroidInfo} from '../../../types/bundle.info';
import {androidFacts} from '../bundleFactsModel';
import {Facts} from './Facts';

export const AndroidSection: React.FC<{ android: AndroidInfo, now: number }> = ({android, now}) =>
    <Section title="Android" description="From the manifest, as aapt reads it" fullHeight={true}>
        <Facts facts={androidFacts(android)} now={now}/>
    </Section>;
