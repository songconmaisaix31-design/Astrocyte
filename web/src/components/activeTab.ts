import { createContext, useContext } from 'react';

export const ActiveTab = createContext('overview');
export function useActiveTab() { return useContext(ActiveTab); }
