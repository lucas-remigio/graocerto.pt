// Helper for email: must have exactly one '@' and at least one '.' after '@'
function isValidEmailStructure(email: string): boolean {
	const at = email.indexOf('@');
	const lastAt = email.lastIndexOf('@');
	const dot = email.lastIndexOf('.');
	return at > 0 && at === lastAt && dot > at + 1 && dot < email.length - 1;
}

export function validateEmail(email: string): boolean {
	email = email.trim();
	return !!email && email.length <= 255 && isValidEmailStructure(email);
}

function isCharacterLowercase(c: string): boolean {
	return c >= 'a' && c <= 'z';
}

function isCharacterUppercase(c: string): boolean {
	return c >= 'A' && c <= 'Z';
}

function isCharacterDigit(c: string): boolean {
	return c >= '0' && c <= '9';
}

function isCharacterSpecial(c: string): boolean {
	const specials = '@$!%*?&#^+=-_.';
	return specials.includes(c);
}

export function isPasswordLengthValid(password: string): boolean {
	password = password.trim();
	return password.length >= 8 && password.length <= 64;
}

export function isPasswordValid(password: string): boolean {
	password = password.trim();
	if (!isPasswordLengthValid(password)) return false;

	let hasUpper = false,
		hasLower = false,
		hasDigit = false,
		hasSpecial = false;

	for (const c of password) {
		if (isCharacterLowercase(c)) hasLower = true;
		else if (isCharacterUppercase(c)) hasUpper = true;
		else if (isCharacterDigit(c)) hasDigit = true;
		else if (isCharacterSpecial(c)) hasSpecial = true;
		if (hasUpper && hasLower && hasDigit && hasSpecial) break;
	}
	return hasUpper && hasLower && hasDigit && hasSpecial;
}
