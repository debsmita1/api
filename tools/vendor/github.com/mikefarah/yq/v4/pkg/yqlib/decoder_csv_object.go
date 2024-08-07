package yqlib

import (
	"encoding/csv"
	"errors"
	"io"

	yaml "gopkg.in/yaml.v3"
)

type csvObjectDecoder struct {
	separator rune
	reader    csv.Reader
	finished  bool
}

func NewCSVObjectDecoder(separator rune) Decoder {
	return &csvObjectDecoder{separator: separator}
}

func (dec *csvObjectDecoder) Init(reader io.Reader) {
	dec.reader = *csv.NewReader(reader)
	dec.reader.Comma = dec.separator
	dec.finished = false
}

func (dec *csvObjectDecoder) convertToYamlNode(content string) *yaml.Node {
	node, err := parseSnippet(content)
	if err != nil {
		return createScalarNode(content, content)
	}
	return node
}

func (dec *csvObjectDecoder) createObject(headerRow []string, contentRow []string) *yaml.Node {
	objectNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}

	for i, header := range headerRow {
		objectNode.Content = append(
			objectNode.Content,
			createScalarNode(header, header),
			dec.convertToYamlNode(contentRow[i]))
	}
	return objectNode
}

func (dec *csvObjectDecoder) Decode(rootYamlNode *yaml.Node) error {
	if dec.finished {
		return io.EOF
	}
	headerRow, err := dec.reader.Read()
	log.Debugf(": headerRow%v", headerRow)
	if err != nil {
		return err
	}

	rootArray := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}

	contentRow, err := dec.reader.Read()

	for err == nil && len(contentRow) > 0 {
		log.Debugf("Adding contentRow: %v", contentRow)
		rootArray.Content = append(rootArray.Content, dec.createObject(headerRow, contentRow))
		contentRow, err = dec.reader.Read()
		log.Debugf("Read next contentRow: %v, %v", contentRow, err)
	}
	if !errors.Is(err, io.EOF) {
		return err
	}

	log.Debugf("finished, contentRow%v", contentRow)
	log.Debugf("err: %v", err)

	rootYamlNode.Kind = yaml.DocumentNode
	rootYamlNode.Content = []*yaml.Node{rootArray}
	return nil
}
