package language

const (
	referenceKey          = "__reference"
	templateReferencesKey = "__references"
)

type Reference map[string]interface{}

func (r Reference) getTarget() string {
	return r[referenceKey].(string)
}

func newReferenceObject(reference string) Reference {
	return Reference{
		referenceKey: reference,
	}
}

// newReferenceEntry
func newReferenceEntry(target string, recipient string) map[string]interface{} {
	return map[string]interface{}{
		"target":    target,
		"recipient": recipient,
	}
}

// findReferencesInTemplate
func findReferencesInTemplate(template *Template, env *environment) []interface{} {
	var references []interface{}
	for target, recipients := range env.referenceRecipients {
		if isReferenceTargetInTemplate(target, template, env) {
			for _, recipient := range recipients {
				references = append(references, newReferenceEntry(target, recipient))
			}
		}
	}

	return references
}

// isReferenceTargetInTemplate
func isReferenceTargetInTemplate(target string, template *Template, env *environment) bool {
	meta := env.getPropertyMetaData(target)
	return meta != nil && meta.template != nil && meta.template == template
}
