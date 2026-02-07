import type { Key } from "../../../types";

export interface KeyStatus {
    isExpired: boolean;
    expiresText: string;
    usagePercent: number;
    isDepleted: boolean;
    isMock: boolean;
    budgetLimit: number;
}

export function getKeyStatus(key: Key): KeyStatus {
    const { expires_at, budget_usage, configuration } = key;
    let isExpired = false;
    let expiresText = "Never";

    if (expires_at) {
        const expDate = new Date(expires_at * 1000);
        expiresText = expDate.toLocaleDateString();
        if (expDate < new Date()) {
            isExpired = true;
        }
    }

    const budgetLimit = configuration?.budget_limit || 0;
    const isMock = configuration?.provider.id === "mock";
    const usagePercent = budgetLimit > 0 ? Math.min((budget_usage / budgetLimit) * 100, 100) : 0;
    const isDepleted = budgetLimit > 0 && budget_usage >= budgetLimit;

    return { isExpired, expiresText, usagePercent, isDepleted, isMock, budgetLimit };
}
