export type CustomerServiceContent = {
    enabled: boolean;
    title: string;
    description: string;
    contactName: string;
    phone: string;
    wechat: string;
};

export type CustomerServiceSetting = CustomerServiceContent & { qrCodeResourceId: string };
export type PublicCustomerService = CustomerServiceContent & { qrCodeUrl: string };

export const DEFAULT_CUSTOMER_SERVICE: CustomerServiceSetting = {
    enabled: false,
    title: "联系客服",
    description: "",
    contactName: "",
    phone: "",
    wechat: "",
    qrCodeResourceId: "",
};

export function validateCustomerService(value: CustomerServiceSetting) {
    for (const [key, label, max] of [
        ["title", "客服弹窗标题", 40],
        ["description", "客服说明", 1000],
        ["contactName", "客服名称", 80],
        ["phone", "客服电话", 80],
        ["wechat", "客服微信号", 80],
    ] as const) {
        const text = value[key].trim();
        if (Array.from(text).length > max || /[\u0000-\u0008\u000B\u000C\u000E-\u001F\u007F]/.test(text)) return `${label}不能超过 ${max} 个字符，且不能包含控制字符`;
    }
    if (value.enabled && !value.title.trim()) return "请填写客服弹窗标题";
    if (value.enabled && ![value.description, value.contactName, value.phone, value.wechat, value.qrCodeResourceId].some((text) => text.trim())) return "启用客服前请至少填写一项内容或上传二维码";
    return "";
}
