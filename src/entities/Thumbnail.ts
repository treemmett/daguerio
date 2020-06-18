import {
  BeforeInsert,
  BeforeUpdate,
  Column,
  CreateDateColumn,
  Entity,
  ManyToOne,
  PrimaryColumn,
} from 'typeorm';
import { Field, ID, Int, ObjectType, registerEnumType } from 'type-graphql';
import {
  IsNotEmpty,
  IsUUID,
  Max,
  MaxLength,
  Min,
  validateOrReject,
} from 'class-validator';
import Photo from './photo';

export enum ThumbnailType {
  NORMAL = 'NORMAL',
  BLUR = 'BLUR',
}

registerEnumType(ThumbnailType, { name: 'ThumbnailType' });

@Entity({ name: 'thumbnails' })
@ObjectType()
export default class Thumbnail {
  @PrimaryColumn('uuid')
  @Field(() => ID)
  @IsUUID('4')
  public readonly id: string;

  @Column({ type: 'int' })
  @Field(() => Int)
  @IsNotEmpty()
  @Min(1)
  @Max(2147483647)
  public size: number;

  @Column({ type: 'smallint' })
  @Field(() => Int)
  @IsNotEmpty()
  @Min(1)
  @Max(32767)
  public width: number;

  @Column({ type: 'smallint' })
  @Field(() => Int)
  @IsNotEmpty()
  @Min(1)
  @Max(32767)
  public height: number;

  @Column({ length: 32 })
  @Field()
  @IsNotEmpty()
  @MaxLength(32)
  public mime: string;

  @Column({ enum: ThumbnailType, type: 'enum' })
  @Field(() => ThumbnailType)
  public type: ThumbnailType;

  @ManyToOne(() => Photo, photo => photo.thumbnails, { onDelete: 'CASCADE' })
  @Field(() => Photo)
  public photo: Promise<Photo>;

  @CreateDateColumn()
  @Field()
  public readonly uploadedTime: Date;

  public constructor(id: string) {
    this.id = id;
  }

  @BeforeInsert()
  @BeforeUpdate()
  public async validate(): Promise<void> {
    await validateOrReject(this);
  }
}
