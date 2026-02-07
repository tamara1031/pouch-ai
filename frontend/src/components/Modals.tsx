import { useState, useEffect, useRef } from "preact/hooks";
import type { Key, MiddlewareInfo, ProviderInfo } from "../types";
import CreateKeyModal from "./modals/CreateKeyModal";
import EditKeyModal from "./modals/EditKeyModal";
import NewKeyDisplayModal from "./modals/NewKeyDisplayModal";

export default function Modals() {
    const [editKey, setEditKey] = useState<Key | null>(null);
    const [newKeyRaw, setNewKeyRaw] = useState<string | null>(null);
    const [middlewareInfo, setMiddlewareInfo] = useState<MiddlewareInfo[]>([]);
    const [providerInfo, setProviderInfo] = useState<ProviderInfo[]>([]);

    const createModalRef = useRef<HTMLInputElement>(null);
    const editModalRef = useRef<HTMLInputElement>(null);
    const newKeyModalRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        const handleOpenEdit = (e: any) => {
            setEditKey(e.detail);
            if (editModalRef.current) editModalRef.current.checked = true;
        };

        const loadInfo = async () => {
            try {
                const [mwRes, pRes] = await Promise.all([
                    fetch("/v1/config/middlewares", { cache: "no-store" }),
                    fetch("/v1/config/providers", { cache: "no-store" })
                ]);
                const mwData = await mwRes.json();
                const pData = await pRes.json();
                setMiddlewareInfo(mwData?.middlewares || []);
                setProviderInfo(pData.providers || []);
            } catch (err) {
                console.error("Failed to load plugin info:", err);
            }
        };

        loadInfo();
        window.addEventListener('open-edit-modal', handleOpenEdit);
        return () => window.removeEventListener('open-edit-modal', handleOpenEdit);
    }, []);

    const handleCreateSuccess = (rawKey: string) => {
        setNewKeyRaw(rawKey);
        if (newKeyModalRef.current) newKeyModalRef.current.checked = true;
    };

    return (
        <>
            <CreateKeyModal
                modalRef={createModalRef}
                onSuccess={handleCreateSuccess}
                middlewareInfos={middlewareInfo}
                providerInfos={providerInfo}
            />
            <EditKeyModal
                modalRef={editModalRef}
                editKey={editKey}
                middlewareInfos={middlewareInfo}
                providerInfos={providerInfo}
            />
            <NewKeyDisplayModal modalRef={newKeyModalRef} newKeyRaw={newKeyRaw} />
        </>
    );
}
