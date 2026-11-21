import React from 'react';
import {Section} from '../../../components/common';
import {IAppBundleDetail} from '../../../types/app';
import {fileFacts} from '../bundleFactsModel';
import {Facts} from './Facts';

export const FileSection: React.FC<{ detail: IAppBundleDetail, now: number }> = ({detail, now}) =>
    <Section title="File" fullHeight={true}>
        <Facts facts={fileFacts(detail, now)} now={now}/>
    </Section>;
