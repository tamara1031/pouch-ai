import { useState, useEffect, useRef } from "preact/hooks";
import type { Key } from "../types";
import CreateKeyModal from "./modals/CreateKeyModal";
import EditKeyModal from "./modals/EditKeyModal";
import NewKeyDisplayModal from "./modals/NewKeyDisplayModal";

export default function Modals() {
    const [editKey, setEditKey] = useState<Key | null>(null);
    const [newKeyRaw, setNewKeyRaw] = useState<string | null>(null);

    const createModalRef = useRef<HTMLInputElement>(null);
    const editModalRef = useRef<HTMLInputElement>(null);
    const newKeyModalRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        const handleOpenEdit = (e: any) => {
            setEditKey(e.detail);
            if (editModalRef.current) editModalRef.current.checked = true;
        };

        window.addEventListener('open-edit-modal', handleOpenEdit);
        return () => window.removeEventListener('open-edit-modal', handleOpenEdit);
    }, []);

    const handleCreateSuccess = (rawKey: string) => {
        setNewKeyRaw(rawKey);
        if (newKeyModalRef.current) newKeyModalRef.current.checked = true;
    };

    return (
        <>
            <CreateKeyModal modalRef={createModalRef} onSuccess={handleCreateSuccess} />
            <EditKeyModal modalRef={editModalRef} editKey={editKey} />
            <NewKeyDisplayModal modalRef={newKeyModalRef} newKeyRaw={newKeyRaw} />
        </>
    );
}
