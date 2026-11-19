import { useEffect, useState, type RefObject } from "react";
import { Form, Input } from "antd";
import { Headset } from "lucide-react";
import { Switch } from "@/pages/admin/ui/controls";
import { SettingsSectionCard } from "@/pages/admin/components/admin-ui";
import { CustomerServiceContent } from "@/components/layout/customer-service-content";
import type { CustomerServiceSetting } from "@/lib/customer-service";
import type { AppearanceAssetSlot } from "@/services/api/appearance";
import { adminResourceFileUrl } from "@/services/api/admin-storage";
import { AssetPicker } from "../appearance-settings-parts";

export function CustomerServiceEditor({
    value,
    onChange,
    file,
    removed,
    inputRef,
    secondFile,
    secondRemoved,
    secondInputRef,
    onSelect,
    onReset,
    disabled,
}: {
    value: CustomerServiceSetting;
    onChange: (value: CustomerServiceSetting) => void;
    file: File | null;
    removed: boolean;
    inputRef: RefObject<HTMLInputElement | null>;
    secondFile: File | null;
    secondRemoved: boolean;
    secondInputRef: RefObject<HTMLInputElement | null>;
    onSelect: (slot: AppearanceAssetSlot, file?: File) => void;
    onReset: (slot: AppearanceAssetSlot) => void;
    disabled: boolean;
}) {
    const objectURL = useFilePreview(file);
    const secondObjectURL = useFilePreview(secondFile);
    const change = (patch: Partial<CustomerServiceSetting>) => onChange({ ...value, ...patch });
    const qrCodeUrl = objectURL || (!removed && value.qrCodeResourceId ? adminResourceFileUrl(value.qrCodeResourceId) : "");
    const secondQrCodeUrl = secondObjectURL || (!secondRemoved && value.secondQrCodeResourceId ? adminResourceFileUrl(value.secondQrCodeResourceId) : "");

    return (
        <SettingsSectionCard className="admin-appearance-section" icon={<Headset className="size-4" aria-hidden="true" />} title="客服配置" description="开启后在工作台右上角显示“客服”。修改完成后点击页面顶部“保存修改”生效；关闭会保留配置。">
            <div className="admin-appearance-section-form">
                <Form className="admin-appearance-form" layout="vertical" disabled={disabled}>
                    <Form.Item label="显示客服入口">
                        <Switch checked={value.enabled} disabled={disabled} aria-label="显示客服入口" onChange={(enabled) => change({ enabled })} />
                    </Form.Item>
                    <Form.Item label="弹窗标题" required={value.enabled} htmlFor="customer-service-title">
                        <Input id="customer-service-title" value={value.title} maxLength={40} onChange={(event) => change({ title: event.target.value })} />
                    </Form.Item>
                    <Form.Item label="说明文案" htmlFor="customer-service-description">
                        <Input.TextArea id="customer-service-description" value={value.description} maxLength={1000} rows={4} showCount placeholder="例如：添加微信，获取使用帮助" onChange={(event) => change({ description: event.target.value })} />
                    </Form.Item>
                    <h3 className="mb-4 font-medium">客服一</h3>
                    <div className="admin-appearance-form-grid">
                        {(
                            [
                                ["contactName", "客服名称"],
                                ["phone", "手机号 / 联系电话"],
                                ["wechat", "微信号"],
                            ] as const
                        ).map(([key, label]) => (
                            <Form.Item key={key} label={label} htmlFor={`customer-service-${key}`}>
                                <Input id={`customer-service-${key}`} value={value[key]} maxLength={80} onChange={(event) => change({ [key]: event.target.value })} />
                            </Form.Item>
                        ))}
                    </div>
                </Form>
                <AssetPicker
                    slot="customer-service-qr"
                    title="客服二维码"
                    description="PNG / JPEG / WebP，最大 5MB，建议使用清晰的正方形图片。"
                    configured={!removed && Boolean(value.qrCodeResourceId)}
                    file={file}
                    inputRef={inputRef}
                    onSelect={onSelect}
                    onReset={onReset}
                    disabled={disabled}
                    emptyLabel="未配置二维码"
                />
                <Form className="admin-appearance-form mt-6" layout="vertical" disabled={disabled}>
                    <h3 className="mb-4 font-medium">客服二（可选）</h3>
                    <div className="admin-appearance-form-grid">
                        {(
                            [
                                ["secondContactName", "第二位客服名称"],
                                ["secondPhone", "第二位手机号 / 联系电话"],
                                ["secondWechat", "第二位微信号"],
                            ] as const
                        ).map(([key, label]) => (
                            <Form.Item key={key} label={label} htmlFor={`customer-service-${key}`}>
                                <Input id={`customer-service-${key}`} value={value[key]} maxLength={80} onChange={(event) => change({ [key]: event.target.value })} />
                            </Form.Item>
                        ))}
                    </div>
                </Form>
                <AssetPicker
                    slot="customer-service-second-qr"
                    title="第二位客服二维码"
                    description="PNG / JPEG / WebP，最大 5MB；留空时不展示第二个二维码。"
                    configured={!secondRemoved && Boolean(value.secondQrCodeResourceId)}
                    file={secondFile}
                    inputRef={secondInputRef}
                    onSelect={onSelect}
                    onReset={onReset}
                    disabled={disabled}
                    emptyLabel="未配置第二位客服二维码"
                />
                <section className="mt-6 rounded-lg border border-border bg-background p-4" aria-label="客服弹窗实时预览">
                    <h3 className="font-medium">{value.title || "联系客服"}</h3>
                    <CustomerServiceContent value={{ ...value, qrCodeUrl, secondQrCodeUrl }} />
                </section>
            </div>
        </SettingsSectionCard>
    );
}

function useFilePreview(file: File | null) {
    const [url, setURL] = useState("");
    useEffect(() => {
        if (!file) {
            setURL("");
            return;
        }
        const nextURL = URL.createObjectURL(file);
        setURL(nextURL);
        return () => URL.revokeObjectURL(nextURL);
    }, [file]);
    return url;
}
