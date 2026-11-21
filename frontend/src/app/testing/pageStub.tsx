import React from 'react';

// A stand-in for a page in routing tests: shows its name, and the `edit` flag
// of the pages that have one.
export const pageStub = (name: string): React.FC<{ edit?: boolean }> => {
    const Stub: React.FC<{ edit?: boolean }> = ({edit}) => <p>{edit === undefined ? name : `${name} edit=${edit}`}</p>;
    return Stub;
};
